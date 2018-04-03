package property

// Property represents a constrained value which can be a Map, a List, or a primitive.
// If the property is a Map, keys must be strings and values are properties.
// These constraints are only enforced by the builder functions: Build, BuildMap, and BuildList.
type Property interface{}
