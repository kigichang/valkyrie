// Package pair provides a Pair[K, V] type holding a key and its associated value,
// the element type of a hash map.
package pair

// Pair holds a key and its associated value.
type Pair[K comparable, V any] struct {
	key   K
	value V
}

// New returns a Pair holding key and value.
func New[K comparable, V any](key K, value V) Pair[K, V] {
	return Pair[K, V]{key: key, value: value}
}

// Key returns the key.
func (p Pair[K, V]) Key() K { return p.key }

// Value returns the value.
func (p Pair[K, V]) Value() V { return p.value }
