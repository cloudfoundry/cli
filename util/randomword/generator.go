package randomword

import (
	"fmt"
	"math/rand"
	"time"
)

type Generator struct{}

func randomInt() int {
	rand.Seed(time.Now().UnixNano())
	// We use 10000 to increase the range of random-ness and minimize potential collisions (with just 1000, we were seeing collisions)
	return rand.Intn(10000)
}

func (Generator) RandomAdjective() string {
	return fmt.Sprintf("adj%d", randomInt())
}

func (Generator) RandomNoun() string {
	return fmt.Sprintf("noun%d", randomInt())
}
