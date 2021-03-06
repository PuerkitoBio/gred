// Package types implements the different value types offered by Redis.
//
// Types are implemented with no knowledge of the concepts of Keys,
// Databases, network connections, or protocols.
package types

// Value is the common interface implemented by all Redis values.
type Value interface {
	Type() string
}

// empty is allocated once and reused by all commands that must return
// an empty list of strings.
var empty = []string{}
