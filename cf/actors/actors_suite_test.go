package actors_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestActors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Actors Suite")
}
