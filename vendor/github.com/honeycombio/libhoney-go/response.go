package libhoney

import (
	"encoding/json"
	"errors"
	"time"
)

// Response is a record of an event sent. It includes information about sending
// the event - how long it took, whether it succeeded, and so on. It also has a
// metadata field that is just a pass through - populate an Event's Metadata
// field and what you put there will be on the Response that corresponds to
// that Event. This allows you to track specific events.
type Response struct {

	// Err contains any error returned by the httpClient on sending or an error
	// indicating queue overflow
	Err error

	// StatusCode contains the HTTP Status Code returned by the Honeycomb API
	// server
	StatusCode int

	// Body is the body of the HTTP response from the Honeycomb API server.
	Body []byte

	// Duration is a measurement of how long the HTTP request to send an event
	// took to process. The actual time it takes libhoney to send an event may
	// be longer due to any time the event spends waiting in the queue before
	// being sent.
	Duration time.Duration

	// Metadata is whatever content you put in the Metadata field of the event for
	// which this is the response. It is passed through unmodified.
	Metadata interface{}
}

func (r *Response) UnmarshalJSON(b []byte) error {
	aux := struct {
		Error  string
		Status int
	}{}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	r.StatusCode = aux.Status
	if aux.Error != "" {
		r.Err = errors.New(aux.Error)
	}
	return nil
}

// writeToResponse adds the response to the response queue. Returns true if it
// dropped the response because it's set to not block on the queue being full
// and the queue was full.
func writeToResponse(resp Response, block bool) (dropped bool) {
	logger.Printf("got response code %d, error %s, and body %s", resp.StatusCode, resp.Err, string(resp.Body))
	if block {
		responses <- resp
	} else {
		select {
		case responses <- resp:
		default:
			return true
		}
	}
	return false
}
