// Package constant contains types and constants used by the ccv3 package.
//
// Constant Naming Conventions:
//
// The standard naming for a constant is <Constant Type><Enum Name>. The only
// exception is 'state' types, where the word 'state' is omitted.
//
// For Example:
//   Constant Type: PackageType
//   Enum Name: Bits
//   Enum Value: "bits"
//   const PackageTypeBits PackageType = "bits"
//
//   Constant Type: PackageState
//   Enum Name: Expired
//   Enum Value: "EXPIRED"
//   const PackageExpired PackageState = "EXPIRED"
package constant
