package set

import "sync"

// SetTS defines a thread safe set data structure.
type SetTS[T comparable] struct {
	set[T]
	l sync.RWMutex // we name it because we don't want to expose it
}

// New creates and initialize a new Set. It's accept a variable number of
// arguments to populate the initial set. If nothing passed a Set with zero
// size is created.
func newTS[T comparable]() *SetTS[T] {
	s := &SetTS[T]{}
	s.m = make(map[T]struct{})

	// Ensure interface compliance
	var _ Set[T] = s

	return s
}

// Add includes the specified items (one or more) to the set. The underlying
// Set s is modified. If passed nothing it silently returns.
func (s *SetTS[T]) Add(items ...T) {
	if len(items) == 0 {
		return
	}

	s.l.Lock()
	defer s.l.Unlock()

	for _, item := range items {
		s.m[item] = keyExists
	}
}

// Remove deletes the specified items from the set.  The underlying Set s is
// modified. If passed nothing it silently returns.
func (s *SetTS[T]) Remove(items ...T) {
	if len(items) == 0 {
		return
	}

	s.l.Lock()
	defer s.l.Unlock()

	for _, item := range items {
		delete(s.m, item)
	}
}

// Pop  deletes and return an item from the set. The underlying Set s is
// modified. If set is empty, nil is returned.
func (s *SetTS[T]) Pop() (T, bool) {
	s.l.RLock()
	for item := range s.m {
		s.l.RUnlock()
		s.l.Lock()
		delete(s.m, item)
		s.l.Unlock()
		return item, true
	}
	s.l.RUnlock()
	var zeroVal T
	return zeroVal, false
}

// Has looks for the existence of items passed. It returns false if nothing is
// passed. For multiple items it returns true only if all of  the items exist.
func (s *SetTS[T]) Has(items ...T) bool {
	// assume checked for empty item, which not exist
	if len(items) == 0 {
		return false
	}

	s.l.RLock()
	defer s.l.RUnlock()

	has := true
	for _, item := range items {
		if _, has = s.m[item]; !has {
			break
		}
	}
	return has
}

// Size returns the number of items in a set.
func (s *SetTS[T]) Size() int {
	s.l.RLock()
	defer s.l.RUnlock()

	l := len(s.m)
	return l
}

// Clear removes all items from the set.
func (s *SetTS[T]) Clear() {
	s.l.Lock()
	defer s.l.Unlock()

	s.m = make(map[T]struct{})
}

// IsEqual test whether s and t are the same in size and have the same items.
func (s *SetTS[T]) IsEqual(t Set[T]) bool {
	s.l.RLock()
	defer s.l.RUnlock()

	// Force locking only if given set is threadsafe.
	if conv, ok := t.(*SetTS[T]); ok {
		conv.l.RLock()
		defer conv.l.RUnlock()
	}

	// return false if they are no the same size
	if sameSize := len(s.m) == t.Size(); !sameSize {
		return false
	}

	equal := true
	t.Each(func(item T) bool {
		_, equal = s.m[item]
		return equal // if false, Each() will end
	})

	return equal
}

// IsSubset tests whether t is a subset of s.
func (s *SetTS[T]) IsSubset(t Set[T]) (subset bool) {
	s.l.RLock()
	defer s.l.RUnlock()

	subset = true

	t.Each(func(item T) bool {
		_, subset = s.m[item]
		return subset
	})

	return
}

// Each traverses the items in the Set, calling the provided function for each
// set member. Traversal will continue until all items in the Set have been
// visited, or if the closure returns false.
func (s *SetTS[T]) Each(f func(item T) bool) {
	s.l.RLock()
	defer s.l.RUnlock()

	for item := range s.m {
		if !f(item) {
			break
		}
	}
}

// List returns a slice of all items. There is also StringSlice() and
// IntSlice() methods for returning slices of type string or int.
func (s *SetTS[T]) List() []T {
	s.l.RLock()
	defer s.l.RUnlock()

	list := make([]T, 0, len(s.m))

	for item := range s.m {
		list = append(list, item)
	}

	return list
}

// Copy returns a new Set with a copy of s.
func (s *SetTS[T]) Copy() Set[T] {
	u := newTS[T]()
	for item := range s.m {
		u.Add(item)
	}
	return u
}

// Merge is like Union, however it modifies the current set it's applied on
// with the given t set.
func (s *SetTS[T]) Merge(t Set[T]) {
	s.l.Lock()
	defer s.l.Unlock()

	t.Each(func(item T) bool {
		s.m[item] = keyExists
		return true
	})
}
