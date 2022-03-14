package stream

import (
	"golang.org/x/exp/slices"
)

type sliceStream[Elem any] struct {
	slice      []Elem
	parallel   bool
	goroutines int
}

// NewSlice new stream instance, generics constraints based on any.
func NewSlice[Elem any](v []Elem) sliceStream[Elem] {
	if v == nil {
		return sliceStream[Elem]{}
	}
	clone := make([]Elem, len(v))
	copy(clone, v)
	return sliceStream[Elem]{slice: clone}
}

// Parallel goroutines > 1 enable parallel, goroutines <= 1 disable parallel
func (stream sliceStream[Elem]) Parallel(goroutines int) sliceStream[Elem] {
	stream.goroutines = goroutines
	if stream.goroutines > 1 {
		stream.parallel = true
	} else {
		stream.parallel = false
	}
	return stream
}

// At Returns the element at the given index. Accepts negative integers, which count back from the last item.
func (stream sliceStream[Elem]) At(index int) Elem {
	l := len(stream.slice)
	if index < 0 {
		index = index + l
	}
	if l == 0 || index < 0 || index >= l {
		var defaultVal Elem
		return defaultVal
	}
	return stream.slice[index]
}

// AllMatch Returns whether all elements in the stream match the provided predicate.
// If the slice is empty or nil then true is returned.
//
// Support Parallel.
func (stream sliceStream[Elem]) AllMatch(predicate func(Elem) bool) bool {
	if stream.parallel {
		handler := func(index int, v Elem) (isReturn bool, taskResult bool) {
			return !predicate(v), false
		}
		return parallelProcess[Elem, bool, bool](
			stream.goroutines,
			stream.slice,
			handler,
			singleResultHandler(true),
			false)
	}

	for _, v := range stream.slice {
		if !predicate(v) {
			return false
		}
	}
	return true

}

// AnyMatch Returns whether any elements in the stream match the provided predicate.
// If the slice is empty or nil then false is returned.
//
// Support Parallel.
func (stream sliceStream[Elem]) AnyMatch(predicate func(Elem) bool) bool {
	if stream.parallel {
		handler := func(index int, v Elem) (isReturn bool, taskResult bool) {
			return predicate(v), true
		}
		return parallelProcess[Elem, bool, bool](
			stream.goroutines,
			stream.slice,
			handler,
			singleResultHandler(false),
			false)
	}

	for _, v := range stream.slice {
		if predicate(v) {
			return true
		}
	}
	return false
}

// Append appends elements to the end of this stream
func (stream sliceStream[Elem]) Append(elements ...Elem) sliceStream[Elem] {
	newSlice := make([]Elem, 0, len(stream.slice)+len(elements))
	newSlice = append(newSlice, stream.slice...)
	newSlice = append(newSlice, elements...)
	stream.slice = newSlice
	return stream
}

// Count Returns the count of elements in this stream.
func (stream sliceStream[Elem]) Count() int {
	return len(stream.slice)
}

// EqualFunc Returns whether the slice in the stream is equal to the destination slice.
// Equal according to the slices.EqualFunc
func (stream sliceStream[Elem]) EqualFunc(dest []Elem, equal func(Elem, Elem) bool) bool {
	return slices.EqualFunc(stream.slice, dest, equal)
}

// ForEach Performs an action for each element of this stream.
//
// Support Parallel.
// Parallel side effects are not executed in the original order of stream elements.
func (stream sliceStream[Elem]) ForEach(action func(int, Elem)) sliceStream[Elem] {
	if stream.slice == nil {
		return stream
	}

	if stream.parallel {
		handler := func(index int, v Elem) (isReturn bool, taskResult Elem) {
			action(index, v)
			return false, taskResult
		}
		parallelProcess[Elem, Elem, []Elem](
			stream.goroutines,
			stream.slice,
			handler,
			multipleResultHandler[Elem](len(stream.slice)),
			true)
		return stream
	}

	for i, v := range stream.slice {
		action(i, v)
	}
	return stream
}

// First Returns the first element in the stream.
// If the slice is empty or nil then Elem Type default value is returned.
func (stream sliceStream[Elem]) First() Elem {
	if len(stream.slice) == 0 {
		var defaultVal Elem
		return defaultVal
	}
	return stream.slice[0]
}

// FindFunc Returns the index of the first element in the stream that matches the provided predicate.
// If not found then -1 is returned.
//
// Support Parallel.
// Parallel side effect is that the element found may not be the first to appear
func (stream sliceStream[Elem]) FindFunc(predicate func(Elem) bool) int {
	if stream.parallel {
		handler := func(index int, v Elem) (isReturn bool, taskResult int) {
			return predicate(v), index
		}
		return parallelProcess[Elem, int, int](
			stream.goroutines,
			stream.slice,
			handler,
			singleResultHandler(-1),
			false)
	}

	for i, v := range stream.slice {
		if predicate(v) {
			return i
		}
	}
	return -1
}

// Filter Returns a stream consisting of the elements of this stream that match the given predicate.
//
// Support Parallel.
func (stream sliceStream[Elem]) Filter(predicate func(Elem) bool) sliceStream[Elem] {
	if stream.slice == nil {
		return stream
	}

	if stream.parallel {
		handler := func(index int, v Elem) (isReturn bool, taskResult Elem) {
			return predicate(v), v
		}

		newSlice := parallelProcess[Elem, Elem, []Elem](
			stream.goroutines,
			stream.slice,
			handler,
			multipleResultHandler[Elem](len(stream.slice)),
			true)
		stream.slice = newSlice
		return stream
	}

	newSlice := make([]Elem, 0)
	for _, v := range stream.slice {
		if predicate(v) {
			newSlice = append(newSlice, v)
		}
	}
	stream.slice = newSlice
	return stream
}

// Insert inserts the values v... into s at index
// If index is out of range then use Append to the end
func (stream sliceStream[Elem]) Insert(index int, elements ...Elem) sliceStream[Elem] {
	if len(stream.slice) <= index {
		return stream.Append(elements...)
	}
	stream.slice = slices.Insert(stream.slice, index, elements...)
	return stream
}

// Delete Removes the elements s[i:j] from this stream, returning the modified stream.
// If the slice is empty or nil then do nothing
func (stream sliceStream[Elem]) Delete(i, j int) sliceStream[Elem] {
	if len(stream.slice) == 0 {
		return stream
	}
	stream.slice = slices.Delete(stream.slice, i, j)
	return stream
}

// IsSortedFunc reports whether x is sorted in ascending order.
// Compare according to the constraints.Ordered.
// If the slice is empty or nil then true is returned.
func (stream sliceStream[Elem]) IsSortedFunc(less func(a, b Elem) bool) bool {
	return slices.IsSortedFunc(stream.slice, less)
}

// Limit Returns a stream consisting of the elements of this stream, truncated to be no longer than maxSize in length.
func (stream sliceStream[Elem]) Limit(maxSize int) sliceStream[Elem] {
	if stream.slice == nil {
		return stream
	}

	newSlice := make([]Elem, 0)
	for i := 0; i < len(stream.slice) && i < maxSize; i++ {
		newSlice = append(newSlice, stream.slice[i])
	}
	stream.slice = newSlice
	return stream
}

// Map Returns a stream consisting of the results of applying the given function to the elements of this stream.
//
// Support Parallel.
func (stream sliceStream[Elem]) Map(mapper func(Elem) Elem) sliceStream[Elem] {
	if stream.slice == nil {
		return stream
	}

	if stream.parallel {
		handler := func(index int, v Elem) (isReturn bool, taskResult Elem) {
			return true, mapper(v)
		}

		newSlice := parallelProcess[Elem, Elem, []Elem](
			stream.goroutines,
			stream.slice,
			handler,
			multipleResultHandler[Elem](len(stream.slice)),
			true)
		stream.slice = newSlice
		return stream
	}

	for i, v := range stream.slice {
		stream.slice[i] = mapper(v)
	}
	return stream
}

// MaxFunc Returns the maximum element of this stream.
// - greater: return a > b
// If the slice is empty or nil then Elem Type default value is returned.
func (stream sliceStream[Elem]) MaxFunc(greater func(a, b Elem) bool) Elem {
	var max Elem
	for i, v := range stream.slice {
		if greater(v, max) || i == 0 {
			max = v
		}
	}
	return max
}

// MinFunc Returns the minimum element of this stream.
// - less: return a < b
// If the slice is empty or nil then Elem Type default value is returned.
func (stream sliceOrderedStream[Elem]) MinFunc(less func(a, b Elem) bool) Elem {
	var min Elem
	for i, v := range stream.slice {
		if less(v, min) || i == 0 {
			min = v
		}
	}
	return min
}

// Reduce Returns a slice consisting of the elements of this stream.
func (stream sliceStream[Elem]) Reduce(accumulator func(Elem, Elem) Elem) Elem {
	var result Elem
	if len(stream.slice) == 0 {
		return result
	}

	for _, v := range stream.slice {
		result = accumulator(result, v)
	}
	return result
}

// SortFunc Returns a sorted stream consisting of the elements of this stream.
// Sorted according to slices.SortFunc.
func (stream sliceStream[Elem]) SortFunc(less func(a, b Elem) bool) sliceStream[Elem] {
	slices.SortFunc(stream.slice, less)
	return stream
}

// SortStableFunc Returns a sorted stream consisting of the elements of this stream.
// Sorted according to slices.SortStableFunc.
func (stream sliceStream[Elem]) SortStableFunc(less func(a, b Elem) bool) sliceStream[Elem] {
	slices.SortStableFunc(stream.slice, less)
	return stream
}

// ToSlice Returns a slice consisting of the elements of this stream.
func (stream sliceStream[Elem]) ToSlice() []Elem {
	return stream.slice
}
