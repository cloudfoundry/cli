package randomword

import (
	"fmt"
	"math/rand"
)

type Generator struct{}

func (Generator) RandomAdjective() string {
	return fmt.Sprintf("adj%d", rand.Intn(1000))
}

func (Generator) RandomNoun() string {
	return fmt.Sprintf("noun%d", rand.Intn(1000))
}
