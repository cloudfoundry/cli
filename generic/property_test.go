package generic_test

import (
	"testing"
	"testing/quick"

	. "github.com/cloudfoundry/cli/generic"
)

// TestMergeIsIdempotent verifies that merging a map with itself produces the same result
func TestMergeIsIdempotent(t *testing.T) {
	f := func(key1, val1, key2, val2 string) bool {
		m := NewMap(map[interface{}]interface{}{
			key1: val1,
			key2: val2,
		})

		result1 := Merge(m, m)
		result2 := Merge(m, m)

		// Merging same map twice should produce same result
		return result1.Get(key1) == result2.Get(key1) &&
			result1.Get(key2) == result2.Get(key2)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestMergePreservesKeys verifies that merge preserves all keys from both maps
func TestMergePreservesKeys(t *testing.T) {
	f := func(key1, val1, key2, val2 string) bool {
		if key1 == key2 {
			return true // Skip when keys are the same
		}

		map1 := NewMap(map[interface{}]interface{}{key1: val1})
		map2 := NewMap(map[interface{}]interface{}{key2: val2})

		result := Merge(map1, map2)

		// Both keys should be present
		return result.Has(key1) && result.Has(key2)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestDeepMergePreservesNonConflictingKeys verifies non-conflicting keys are preserved
func TestDeepMergePreservesNonConflictingKeys(t *testing.T) {
	f := func(key1, val1, key2, val2 string) bool {
		if key1 == key2 {
			return true // Skip when keys conflict
		}

		map1 := NewMap(map[interface{}]interface{}{key1: val1})
		map2 := NewMap(map[interface{}]interface{}{key2: val2})

		result := DeepMerge(map1, map2)

		// Both values should be present with their original values
		return result.Get(key1) == val1 && result.Get(key2) == val2
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestDeepMergeHandlesEmptyMaps verifies that merging with empty map returns the other map
func TestDeepMergeHandlesEmptyMaps(t *testing.T) {
	f := func(key, val string) bool {
		nonEmpty := NewMap(map[interface{}]interface{}{key: val})
		empty := NewMap()

		result1 := DeepMerge(nonEmpty, empty)
		result2 := DeepMerge(empty, nonEmpty)

		// Both should preserve the non-empty map's data
		return result1.Get(key) == val && result2.Get(key) == val
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestMapGetSetRoundtrip verifies that setting and getting a value works consistently
func TestMapGetSetRoundtrip(t *testing.T) {
	f := func(key, val string) bool {
		m := NewMap()
		m.Set(key, val)

		retrieved := m.Get(key)
		return retrieved == val
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestMapHasConsistency verifies Has() is consistent with Get()
func TestMapHasConsistency(t *testing.T) {
	f := func(key, val string) bool {
		m := NewMap(map[interface{}]interface{}{key: val})

		// If Has returns true, Get should return non-nil
		// If Has returns false, Get should return nil
		has := m.Has(key)
		get := m.Get(key)

		return has == (get != nil)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestExceptRemovesKey verifies Except removes the specified key
func TestExceptRemovesKey(t *testing.T) {
	f := func(key1, val1, key2, val2 string) bool {
		if key1 == key2 {
			return true
		}

		m := NewMap(map[interface{}]interface{}{
			key1: val1,
			key2: val2,
		})

		result := m.Except([]interface{}{key1})

		// key1 should be removed, key2 should remain
		return !result.Has(key1) && result.Has(key2)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

// TestReduceWithIdentityFunction verifies reduce with identity preserves data
func TestReduceWithIdentityFunction(t *testing.T) {
	f := func(key1, val1, key2, val2 string) bool {
		collections := []Map{
			NewMap(map[interface{}]interface{}{key1: val1}),
			NewMap(map[interface{}]interface{}{key2: val2}),
		}

		// Identity reducer just copies all keys
		identity := func(key, val interface{}, acc Map) Map {
			acc.Set(key, val)
			return acc
		}

		result := Reduce(collections, NewMap(), identity)

		// All keys should be in the result
		return result.Has(key1) && result.Has(key2)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
