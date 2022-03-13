package stream

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func newArray(count int) []int {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	s := make([]int, count)
	for i := 0; i < count; i++ {
		s[i] = r.Intn(count * 2)
	}
	return s
}

func TestParallelFilter(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(v int) bool
		want      int
	}{
		{
			name:      "match",
			input:     newArray(100),
			predicate: func(v int) bool { return v < 100 },
		},
		{
			name:      "match",
			input:     newArray(200),
			predicate: func(v int) bool { return v < 100 },
		},
		{
			name:      "match",
			input:     newArray(300),
			predicate: func(v int) bool { return v < 100 },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t,
				NewSliceByOrdered(tt.input).Parallel(10).Filter(tt.predicate).ToSlice(),
				NewSliceByOrdered(tt.input).Filter(tt.predicate).ToSlice())
		})
	}
}

func TestParallelMap(t *testing.T) {
	tests := []struct {
		name   string
		input  []int
		mapper func(int) int
		want   []int
	}{
		{
			name:   "normal",
			input:  newArray(100),
			mapper: func(i int) int { return i * 2 },
		},
		{
			name:   "empty",
			input:  newArray(100),
			mapper: func(i int) int { return i * 2 },
		},
		{
			name:   "nil",
			input:  newArray(100),
			mapper: func(i int) int { return i * 2 },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t,
				NewSliceByOrdered(tt.input).Parallel(10).Map(tt.mapper).ToSlice(),
				NewSliceByOrdered(tt.input).Map(tt.mapper).ToSlice())
		})
	}
}

func TestEcho(t *testing.T) {
	s := newArray(100)
	NewSliceByOrdered(s).Parallel(10).Filter(func(i int) bool { return i > 100 })
	NewSliceByOrdered(s).Parallel(10).ForEach(func(i int, v int) {})
}
