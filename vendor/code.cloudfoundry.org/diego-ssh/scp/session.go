package scp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"code.cloudfoundry.org/lager"
)

const (
	NEWLINE = "\n"
	SPACE   = " "
)

type Session struct {
	stdin  *bufio.Reader
	stdout io.Writer
	stderr io.Writer

	preserveTimesAndMode bool

	logger lager.Logger
}

func NewSession(stdin io.Reader, stdout io.Writer, stderr io.Writer, preserveTimesAndMode bool, logger lager.Logger) *Session {
	return &Session{
		stdin:                bufio.NewReader(stdin),
		stdout:               stdout,
		stderr:               stderr,
		preserveTimesAndMode: preserveTimesAndMode,
		logger:               logger.Session("scp-session"),
	}
}

func (sess *Session) sendConfirmation() error {
	_, err := sess.stdout.Write([]byte{0})
	return err
}

func (sess *Session) sendWarning(message string) error {
	_, err := sess.stdout.Write([]byte{1})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(sess.stdout, "%s\n", message)
	if err != nil {
		return err
	}

	return nil
}

func (sess *Session) sendError(message string) error {
	_, err := sess.stdout.Write([]byte{1})
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(sess.stdout, "scp: %s\n", message)
	if err != nil {
		return err
	}

	return nil
}

func (sess *Session) awaitConfirmation() error {
	ackType, err := sess.readByte()
	if err != nil {
		return err
	}

	switch ackType {
	case 0:
	case 1:
		message, err := sess.readString(NEWLINE)
		if err != nil {
			return err
		}
		fmt.Fprintf(sess.stderr, message)
	case 2:
		message, err := sess.readString(NEWLINE)
		if err != nil {
			return err
		}
		return errors.New(message)
	default:
		return fmt.Errorf("invalid acknowledgement identifier: %x", ackType)
	}

	return nil
}

func (sess *Session) readString(delim string) (string, error) {
	message, err := sess.stdin.ReadString(delim[0])
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(message, delim), nil
}

func (sess *Session) readByte() (byte, error) {
	message := make([]byte, 1)

	var n int
	var err error
	for n == 0 && err == nil {
		n, err = sess.stdin.Read(message)
	}

	if err != nil {
		return 0, err
	}

	if n != 1 {
		return 0, errors.New("read failed")
	}

	return message[0], nil
}

func (sess *Session) peekByte() (byte, error) {
	b, err := sess.readByte()
	if err != nil {
		return b, err
	}

	err = sess.stdin.UnreadByte()
	if err != nil {
		return b, err
	}

	return b, nil
}
