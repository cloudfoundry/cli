// Package signature signs and validates dropsonde messages.

// Messages are prepended with a HMAC SHA256 signature (the signature makes up
// the first 32 bytes of a signed message; the remainder is the original message
// in cleartext).
package signature

import (
	"crypto/hmac"
	"crypto/sha256"

	"github.com/cloudfoundry/dropsonde/metrics"
	"github.com/cloudfoundry/gosteno"
)

const SIGNATURE_LENGTH = 32

// A SignatureVerifier is a self-instrumenting pipeline object that validates
// and removes signatures.
type Verifier struct {
	logger       *gosteno.Logger
	sharedSecret string
}

// NewSignatureVerifier returns a SignatureVerifier with the provided logger and
// shared signing secret.
func NewVerifier(logger *gosteno.Logger, sharedSecret string) *Verifier {
	return &Verifier{
		logger:       logger,
		sharedSecret: sharedSecret,
	}
}

// Run validates signatures. It consumes signed messages from inputChan,
// verifies the signature, and sends the message (sans signature) to outputChan.
// Invalid messages are dropped and nothing is sent to outputChan. Thus a reader
// of outputChan is guaranteed to receive only messages with a valid signature.
//
// Run blocks on sending to outputChan, so the channel must be drained for the
// function to continue consuming from inputChan.
func (v *Verifier) Run(inputChan <-chan []byte, outputChan chan<- []byte) {
	for signedMessage := range inputChan {
		if len(signedMessage) < SIGNATURE_LENGTH {
			v.logger.Warn("signatureVerifier: missing signature")
			metrics.BatchIncrementCounter("signatureVerifier.missingSignatureErrors")
			continue
		}

		signature, message := signedMessage[:SIGNATURE_LENGTH], signedMessage[SIGNATURE_LENGTH:]
		if v.verifyMessage(message, signature) {
			outputChan <- message
			metrics.BatchIncrementCounter("signatureVerifier.validSignatures")
		} else {
			v.logger.Warn("signatureVerifier: invalid signature")
			metrics.BatchIncrementCounter("signatureVerifier.invalidSignatureErrors")
		}
	}
}

func (v *Verifier) verifyMessage(message, signature []byte) bool {
	expectedMAC := generateSignature(message, []byte(v.sharedSecret))
	return hmac.Equal(signature, expectedMAC)
}

// SignMessage returns a message signed with the provided secret, with the
// signature prepended to the original message.
func SignMessage(message, secret []byte) []byte {
	signature := generateSignature(message, secret)
	return append(signature, message...)
}

func generateSignature(message, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(message)
	return mac.Sum(nil)
}
