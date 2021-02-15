[![test](https://github.com/cloudfoundry/jsonry/workflows/test/badge.svg?branch=main)](https://github.com/cloudfoundry/jsonry/actions?query=workflow%3Atest+branch%3Amain)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/code.cloudfoundry.org/jsonry?tab=doc)

# JSONry

A Go library and notation for converting between a Go `struct` and JSON.

```go
import "code.cloudfoundry.org/jsonry"

s := struct {
  GUID string `jsonry:"relationships.space.data.guid"`
}{
  GUID: "267758c0-985b-11ea-b9ac-48bf6bec2d78",
}

json, _ := jsonry.Marshal(s)
fmt.Println(string(json))
```
Will generate the following JSON:
```json
{
  "relationships": {
    "space": {
      "data": {
        "guid": "267758c0-985b-11ea-b9ac-48bf6bec2d78"
      }
    }
  }
}
```
The operation is reversible using `Unmarshal()`. The key advantage is that nested JSON can be generated and parsed without
the need to create intermediate Go structures. Check out [the documentation](https://pkg.go.dev/code.cloudfoundry.org/jsonry?tab=doc) for details.

JSONry started life in the [Cloud Foundry CLI](https://github.com/cloudfoundry/cli) project. It has been extracted so
that it can be used in other projects too.

More information:
- [License](./LICENSE)
- [Releasing](./RELEASING.md)
- [Performance](./PERFORMANCE.md)
