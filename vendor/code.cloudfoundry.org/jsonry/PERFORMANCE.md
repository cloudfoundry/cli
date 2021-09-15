## Performance

There is a benchmark test which compares using JSONry and `encoding/json` to
marshal and unmarshal the same data. The value of the benchmark is in the
relative performance between the two. In the benchamark test:

- Unmarshal
  - JSONry takes 2.5 times as long as `encoding/json`
  - JSONry allocates 6 times as much memory as `enconding/json`

- Marshal
  - JSONry takes 11 times as long as `encoding/json`
  - JSONry allocates 21 times as much memory as `enconding/json`

- Version: `go version go1.14.3 darwin/amd64`
- Command: `go test -run none -bench . -benchmem -benchtime 10s`
