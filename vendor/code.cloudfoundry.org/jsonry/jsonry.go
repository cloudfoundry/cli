// JSONry exists to make is easier to marshal and unmarshal nested JSON into Go structures.
// For example:
//
//   {
//     "relationships": {
//       "space": {
//         "data": {
//           "guid": "267758c0-985b-11ea-b9ac-48bf6bec2d78"
//         }
//       }
//     }
//   }
//
// Can be marshaled and unmarshaled using the Go struct:
//
//   type s struct { GUID string `jsonry:"relationships.space.data.guid"` }
//
// There is no need to create intermediate structures as would be required with the standard Go JSON parser.
//
// JSONry uses the standard Go parser under the covers.
// JSONry is a trade-off that delivers usability at a slight cost to performance.
//
// JSONry uses struct tags to specify locations in the JSON document.
//
// When there is neither a "jsonry" or "json" struct tag, the JSON object key name
// will be the same as the struct field name. For example:
//  Go: s := struct { Foo string }{Foo: "value"}
//  JSON: {"Foo": "value"}
//
// When there is a "json" struct tag, it will be interpreted in the same way as the standard
// Go JSON parser would interpret it. For example:
//  Go: s := struct { Foo string `json:"bar"` }{Foo: "value"}
//  JSON: {"bar": "value"}
//
// When there is a "jsonry" struct tag, in addition to working like the "json" tag, it may
// also contain "." (period) to denote a nesting hierarchy.  For example:
//  Go: s := struct { Foo string `jsonry:"foo.bar"` }{Foo: "value"}
//  JSON: {"foo": {"bar": "value"} }
//
package jsonry
