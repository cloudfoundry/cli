package cf

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRouteURL(t *testing.T) {
	route := Route{
		Host:   "foo",
		Domain: Domain{Name: "example.com"},
	}

	assert.Equal(t, route.URL(), "foo.example.com")
}

func TestRouteURLWithoutHost(t *testing.T) {
	route := Route{
		Host:   "",
		Domain: Domain{Name: "example.com"},
	}

	assert.Equal(t, route.URL(), "example.com")
}
