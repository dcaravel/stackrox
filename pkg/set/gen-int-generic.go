// Code generated by genny. DO NOT EDIT.
// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/mauricelam/genny

package set

import (
	"github.com/deckarep/golang-set"
)

// If you want to add a set for your custom type, simply add another go generate line along with the
// existing ones. If you're creating a set for a primitive type, you can follow the example of "string"
// and create the generated file in this package.
// Sometimes, you might need to create it in the same package where it is defined to avoid import cycles.
// The permission set is an example of how to do that.
// You can also specify the -imp command to specify additional imports in your generated file, if required.

// int represents a generic type that we want to have a set of.

// IntSet will get translated to generic sets.
// It uses mapset.Set as the underlying implementation, so it comes with a bunch
// of utility methods, and is thread-safe.
type IntSet struct {
	underlying mapset.Set
}

// Add adds an element of type int.
func (k IntSet) Add(i int) bool {
	if k.underlying == nil {
		k.underlying = mapset.NewSet()
	}

	return k.underlying.Add(i)
}

// Remove removes an element of type int.
func (k IntSet) Remove(i int) {
	if k.underlying != nil {
		k.underlying.Remove(i)
	}
}

// Contains returns whether the set contains an element of type int.
func (k IntSet) Contains(i int) bool {
	if k.underlying != nil {
		return k.underlying.Contains(i)
	}
	return false
}

// Cardinality returns the number of elements in the set.
func (k IntSet) Cardinality() int {
	if k.underlying != nil {
		return k.underlying.Cardinality()
	}
	return 0
}

// Difference returns a new set with all elements of k not in other.
func (k IntSet) Difference(other IntSet) IntSet {
	if k.underlying == nil {
		return IntSet{underlying: other.underlying}
	} else if other.underlying == nil {
		return IntSet{underlying: k.underlying}
	}

	return IntSet{underlying: k.underlying.Difference(other.underlying)}
}

// Intersect returns a new set with the intersection of the members of both sets.
func (k IntSet) Intersect(other IntSet) IntSet {
	if k.underlying != nil && other.underlying != nil {
		return IntSet{underlying: k.underlying.Intersect(other.underlying)}
	}
	return IntSet{}
}

// Union returns a new set with the union of the members of both sets.
func (k IntSet) Union(other IntSet) IntSet {
	if k.underlying == nil {
		return IntSet{underlying: other.underlying}
	} else if other.underlying == nil {
		return IntSet{underlying: k.underlying}
	}

	return IntSet{underlying: k.underlying.Union(other.underlying)}
}

// AsSlice returns a slice of the elements in the set. The order is unspecified.
func (k IntSet) AsSlice() []int {
	if k.underlying == nil {
		return nil
	}
	elems := make([]int, 0, k.Cardinality())
	for elem := range k.underlying.Iter() {
		elems = append(elems, elem.(int))
	}
	return elems
}

// NewIntSet returns a new set with the given key type.
func NewIntSet(initial ...int) IntSet {
	k := IntSet{underlying: mapset.NewSet()}
	for _, elem := range initial {
		k.Add(elem)
	}
	return k
}
