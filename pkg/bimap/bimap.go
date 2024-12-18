package bimap

import "fmt"

// BiMap represents a bidirectional map
type BiMap[K comparable, V comparable] struct {
	keyToValue map[K]V
	valueToKey map[V]K
}

// NewBiMap creates a new BiMap
func NewBiMap[K comparable, V comparable]() *BiMap[K, V] {
	return &BiMap[K, V]{
		keyToValue: make(map[K]V),
		valueToKey: make(map[V]K),
	}
}

// NewBiMapFromMap NewBiMap creates a new BiMap from a map.
func NewBiMapFromMap[K comparable, V comparable](m map[K]V) *BiMap[K, V] {
	biMap := NewBiMap[K, V]()

	// Populate the BiMap
	for k, v := range m {
		if _, exists := biMap.valueToKey[v]; exists {
			panic(fmt.Sprintf("Duplicate value %v found for key %v", v, k))
		}

		biMap.keyToValue[k] = v
		biMap.valueToKey[v] = k
	}

	return biMap
}

// Put adds a key-value pair to the BiMap
func (b *BiMap[K, V]) Put(key K, value V) {
	// Remove previous mappings for key and value if they exist
	if oldValue, exists := b.keyToValue[key]; exists {
		delete(b.valueToKey, oldValue)
	}
	if oldKey, exists := b.valueToKey[value]; exists {
		delete(b.keyToValue, oldKey)
	}

	// Insert new mappings
	b.keyToValue[key] = value
	b.valueToKey[value] = key
}

// GetByKey retrieves the value for a given key
func (b *BiMap[K, V]) GetByKey(key K) (V, bool) {
	val, exists := b.keyToValue[key]
	return val, exists
}

// GetByValue retrieves the key for a given value
func (b *BiMap[K, V]) GetByValue(value V) (K, bool) {
	key, exists := b.valueToKey[value]
	return key, exists
}

// GetKeys returns all keys of the BiMap
func (b *BiMap[K, V]) GetKeys() []K {
	var keys []K
	for key := range b.keyToValue {
		keys = append(keys, key)
	}

	return keys
}

// GetValues returns all the values of the BiMap
func (b *BiMap[K, V]) GetValues() []V {
	var values []V
	for val := range b.valueToKey {
		values = append(values, val)
	}

	return values
}

// Delete removes a key-value pair
func (b *BiMap[K, V]) Delete(key K) {
	if value, exists := b.keyToValue[key]; exists {
		delete(b.keyToValue, key)
		delete(b.valueToKey, value)
	}
}
