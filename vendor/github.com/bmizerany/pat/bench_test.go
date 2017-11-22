package pat

import (
	"net/http"
	"testing"
)

func BenchmarkPatternMatching(b *testing.B) {
	p := New()
	p.Get("/hello/:name", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		r := newRequest("GET", "/hello/blake", nil)
		b.StartTimer()
		p.ServeHTTP(nil, r)
	}
}
