package collections

import (
	"cmp"
	"slices"
)

// SortedList brings a bit of high-level programming to slice.
// Should not be used in hotpaths! It's around 25% slower than using slice directly.
type SortedList[T cmp.Ordered] struct {
	// As generic access may be slower, underlying slice is exposed publicly for access in hot paths
	Slice []T
}

func NewSortedList[T cmp.Ordered]() *SortedList[T] {
	return &SortedList[T]{
		Slice: make([]T, 0),
	}
}

func (list *SortedList[T]) Clone() *SortedList[T] {
	return list.CloneWithAddCap(0)
}

func (list *SortedList[T]) CloneWithAddCap(cap int) *SortedList[T] {
	cloned := &SortedList[T]{
		Slice: make([]T, len(list.Slice), len(list.Slice)+cap),
	}

	copy(cloned.Slice, list.Slice)

	return cloned
}

// Add adds item T to the list.
// If comparer has been set, it puts the item in a place defined by comparer's criteria.
// Returns index at which the item has been added.
func (list *SortedList[T]) Add(v T) int {
	n, _ := slices.BinarySearch(list.Slice, v)

	oldLen := len(list.Slice)

	list.Slice = append(list.Slice, v)

	if n != oldLen {
		copy(list.Slice[n+1:], list.Slice[n:])
		list.Slice[n] = v
	}

	return n
}

// AddAllV is the same as AddAll but input is variadic.
func (list *SortedList[T]) AddAllV(v ...T) []int {
	return list.AddAll(v)
}

// AddAll adds items contained in []T to the list.
// If comparer has been set, it puts the items in a place defined by comparer's criteria.
// Returns indices at which items have been added.
func (list *SortedList[T]) AddAll(v []T) (idx []int) {
	for _, v1 := range v {
		idx = append(idx, list.Add(v1))
	}

	return
}

// Clear clears the list.
func (list *SortedList[T]) Clear() {
	list.Slice = list.Slice[:0]
}

// Contains checks if v is present in the list.
func (list *SortedList[T]) Contains(v T) bool {
	return list.Index(v) > -1
}

// First returns the first item in the list.
func (list *SortedList[T]) First() T {
	return list.Slice[0]
}

// Get returns the item at index i.
func (list *SortedList[T]) Get(i int) T {
	return list.Slice[i]
}

// Index returns the index of v, -1 if not found.
func (list *SortedList[T]) Index(v T) int {
	return slices.IndexFunc(list.Slice, func(b T) bool {
		return v == b
	})
}

// IsEmpty returns true if list is empty.
func (list *SortedList[T]) IsEmpty() bool {
	return len(list.Slice) == 0
}

// Last returns the last item in the list.
func (list *SortedList[T]) Last() T {
	return list.Slice[len(list.Slice)-1]
}

// Len returns the size of underlying slice.
func (list *SortedList[T]) Len() int {
	return len(list.Slice)
}

// Remove deletes first encountered v and returns its index before removal. If v has not been found, -1 is returned.
func (list *SortedList[T]) Remove(v T) int {
	i := list.Index(v)

	if i > -1 {
		copy(list.Slice[i:], list.Slice[i+1:])
		list.Slice = list.Slice[:len(list.Slice)-1]
	}

	return i
}

// RemoveAt deletes an item at i and returns its value before removal.
func (list *SortedList[T]) RemoveAt(i int) (v T) {
	v = list.Slice[i]

	copy(list.Slice[i:], list.Slice[i+1:])
	list.Slice = list.Slice[:len(list.Slice)-1]

	return
}

// RemoveRange deletes items in [i, j) range
func (list *SortedList[T]) RemoveRange(i, j int) {
	if i <= 0 {
		list.Slice = list.Slice[j:]
	} else if j >= len(list.Slice) {
		list.Slice = list.Slice[:i]
	} else {
		copy(list.Slice[i:], list.Slice[j:])
		list.Slice = list.Slice[:len(list.Slice)-j+i]
	}
}

// Sort is faster than SortStable, but does not guarantee original order of equal values.
// If stability is critical, add additional criteria (like original index) or use SortStable.
func (list *SortedList[T]) Sort() {
	slices.Sort(list.Slice)
}

// SortF is faster than SortStableF, but does not guarantee original order of equal values.
// If stability is critical, add additional criteria (like original index) or use SortStableF.
func (list *SortedList[T]) SortF(comparer func(a, b T) int) {
	slices.SortFunc(list.Slice, comparer)
}

// SortStableF is slower than SortF, but guarantees original order of equal values.
func (list *SortedList[T]) SortStableF(comparer func(a, b T) int) {
	slices.SortStableFunc(list.Slice, comparer)
}
