package generator_test

import (
	"github.com/cloudfoundry/cli/words/generator"
	"testing"
)

func BenchmarkWordGenerator_Babble(b *testing.B) {
	wordGen := generator.NewWordGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wordGen.Babble()
	}
}

func BenchmarkWordGenerator_NewWordGenerator(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generator.NewWordGenerator()
	}
}

func BenchmarkWordGenerator_MultipleBabbles(b *testing.B) {
	wordGen := generator.NewWordGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10; j++ {
			wordGen.Babble()
		}
	}
}

func BenchmarkWordGenerator_ParallelBabble(b *testing.B) {
	wordGen := generator.NewWordGenerator()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wordGen.Babble()
		}
	})
}
