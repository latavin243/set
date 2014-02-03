package set

import (
	"fmt"
	"strings"
)

// Provides a common set baseline for both threadsafe and non-ts Sets.
type set struct {
	m map[interface{}]struct{} // struct{} doesn't take up space
}

// SetNonTS defines a non-thread safe set data structure.
type SetNonTS struct {
	set
}

// NewNonTS creates and initialize a new non-threadsafe Set.
// It accepts a variable number of arguments to populate the initial set.
// If nothing is passed a SetNonTS with zero size is created.
func NewNonTS(items ...interface{}) *SetNonTS {
	s := &SetNonTS{}
	s.m = make(map[interface{}]struct{})

	s.Add(items...)
	return s
}

// New creates and initalizes a new Set interface. It accepts a variable
// number of arguments to populate the initial set. If nothing is passed a
// zero size Set based on the struct is created.
func (s *set) New(items ...interface{}) Interface {
	return NewNonTS(items...)
}

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s *set) Add(items ...interface{}) {
	if len(items) == 0 {
		return
	}

	for _, item := range items {
		s.m[item] = keyExists
	}
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s *set) Remove(items ...interface{}) {
	if len(items) == 0 {
		return
	}

	for _, item := range items {
		delete(s.m, item)
	}
}

// Pop  deletes and return an item from the set. The underlying Set s is
// modified. If set is empty, nil is returned.
func (s *set) Pop() interface{} {
	for item := range s.m {
		delete(s.m, item)
		return item
	}
	return nil
}

// Has looks for the existence of items passed. It returns false if nothing is
// passed. For multiple items it returns true only if all of  the items exist.
func (s *set) Has(items ...interface{}) bool {
	// assume checked for empty item, which not exist
	if len(items) == 0 {
		return false
	}

	has := true
	for _, item := range items {
		if _, has = s.m[item]; !has {
			break
		}
	}
	return has
}

// Size returns the number of items in a set.
func (s *set) Size() int {
	return len(s.m)
}

// Clear removes all items from the set.
func (s *set) Clear() {
	s.m = make(map[interface{}]struct{})
}

// IsEmpty reports whether the Set is empty.
func (s *set) IsEmpty() bool {
	return s.Size() == 0
}

// IsEqual test whether s and t are the same in size and have the same items.
func (s *set) IsEqual(t Interface) bool {
	// Force locking only if given set is threadsafe.
	if conv, ok := t.(*Set); ok {
		conv.l.RLock()
		defer conv.l.RUnlock()
	}

	equal := true
	if equal = len(s.m) == t.Size(); equal {
		t.Each(func(item interface{}) (equal bool) {
			_, equal = s.m[item]
			return
		})
	}

	return equal
}

// IsSubset tests whether t is a subset of s.
func (s *set) IsSubset(t Interface) (subset bool) {
	subset = true

	t.Each(func(item interface{}) bool {
		_, subset = s.m[item]
		return subset
	})

	return
}

// IsSuperset tests whether t is a superset of s.
func (s *set) IsSuperset(t Interface) bool {
	return t.IsSubset(s)
}

// Each traverses the items in the Set, calling the provided function for each
// set member. Traversal will continue until all items in the Set have been
// visited, or if the closure returns false.
func (s *set) Each(f func(item interface{}) bool) {
	for item := range s.m {
		if !f(item) {
			break
		}
	}
}

// String returns a string representation of s
func (s *set) String() string {
	t := make([]string, 0, len(s.List()))
	for _, item := range s.List() {
		t = append(t, fmt.Sprintf("%v", item))
	}

	return fmt.Sprintf("[%s]", strings.Join(t, ", "))
}

// List returns a slice of all items. There is also StringSlice() and
// IntSlice() methods for returning slices of type string or int.
func (s *set) List() []interface{} {
	list := make([]interface{}, 0, len(s.m))

	for item := range s.m {
		list = append(list, item)
	}

	return list
}

// Copy returns a new Set with a copy of s.
func (s *set) Copy() Interface {
	return NewNonTS(s.List()...)
}

// Union is the merger of two sets. It returns a new set with the element in s
// and t combined. It doesn't modify s. Use Merge() if  you want to change the
// underlying set s.
func (s *set) Union(t Interface) Interface {
	u := s.Copy()
	u.Merge(t)
	return u
}

// Merge is like Union, however it modifies the current set it's applied on
// with the given t set.
func (s *set) Merge(t Interface) {
	t.Each(func(item interface{}) bool {
		s.m[item] = keyExists
		return true
	})
}

// it's not the opposite of Merge.
// Separate removes the set items containing in t from set s. Please aware that
func (s *set) Separate(t Interface) {
	s.Remove(t.List()...)
}

// Intersection returns a new set which contains items which is in both s and t.
func (s *set) Intersection(t Interface) Interface {
	u := s.Copy()
	u.Separate(u.Difference(t))
	return u
}

// Difference returns a new set which contains items which are in s but not in t.
func (s *set) Difference(t Interface) Interface {
	u := s.Copy()
	u.Separate(t)
	return u
}

// SymmetricDifference returns a new set which s is the difference of items which are in
// one of either, but not in both.
func (s *set) SymmetricDifference(t Interface) Interface {
	u := s.Difference(t)
	v := t.Difference(s)
	return u.Union(v)
}

// StringSlice is a helper function that returns a slice of strings of s. If
// the set contains mixed types of items only items of type string are returned.
func (s *set) StringSlice() []string {
	slice := make([]string, 0)
	for _, item := range s.List() {
		v, ok := item.(string)
		if !ok {
			continue
		}

		slice = append(slice, v)
	}
	return slice
}

// IntSlice is a helper function that returns a slice of ints of s. If
// the set contains mixed types of items only items of type int are returned.
func (s *set) IntSlice() []int {
	slice := make([]int, 0)
	for _, item := range s.List() {
		v, ok := item.(int)
		if !ok {
			continue
		}

		slice = append(slice, v)
	}
	return slice
}
