package cf

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRouteURL(t *testing.T) {
	route := Route{}
	route.Host = "foo"

	domain := DomainFields{}
	domain.Name = "example.com"
	route.Domain = domain

	assert.Equal(t, route.URL(), "foo.example.com")
}

func TestRouteURLWithoutHost(t *testing.T) {
	route := Route{}
	route.Host = ""

	domain := DomainFields{}
	domain.Name = "example.com"
	route.Domain = domain

	assert.Equal(t, route.URL(), "example.com")
}
