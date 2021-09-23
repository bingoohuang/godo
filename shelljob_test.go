package main

import (
	"github.com/go-playground/assert/v2"
	"math/rand"
	"testing"
)

type ArrayIter struct {
	State []string
	Pos   int
}

func (a *ArrayIter) HasNext() bool {
	return a.Pos < len(a.State)
}

func (a *ArrayIter) Next()         { a.Pos++ }
func (a *ArrayIter) Value() string { return a.State[a.Pos] }
func (a *ArrayIter) Reset()        { a.Pos = 0 }

var _ Iterator = (*ArrayIter)(nil)

func TestArrange(t *testing.T) {
	out := make(chan struct{})
	wait := make(chan struct{})
	a := &ArrayIter{State: []string{"a", "b", "c"}}
	go Arrange([]Iterator{a}, out, wait)

	var arr []string
	for range out {
		arr = append(arr, a.Value())
		wait <- struct{}{}
	}

	assert.Equal(t, []string{"a", "b", "c"}, arr)

	b := &ArrayIter{State: []string{"1", "2"}}
	out = make(chan struct{}, 0)
	go Arrange([]Iterator{a, b}, out, wait)

	arr = arr[:0]
	for range out {
		arr = append(arr, a.Value()+b.Value())
		wait <- struct{}{}
	}

	assert.Equal(t, []string{"a1", "a2", "b1", "b2", "c1", "c2"}, arr)

	c := &ArrayIter{State: []string{"x", "y"}}
	out = make(chan struct{}, 0)
	go Arrange([]Iterator{a, b, c}, out, wait)

	arr = arr[:0]
	for range out {
		arr = append(arr, a.Value()+b.Value()+c.Value())
		wait <- struct{}{}
	}

	assert.Equal(t, []string{"a1x", "a1y", "a2x", "a2y", "b1x", "b1y", "b2x", "b2y", "c1x", "c1y", "c2x", "c2y"}, arr)
}

func TestArrangement(t *testing.T) {
	assert.Equal(t, [][]string{
		{"a"}, {"b"},
	}, Arrangement([][]string{
		{"a", "b"},
	}))

	assert.Equal(t, [][]string{
		{"a", "1"},
		{"a", "2"},
		{"a", "3"},
		{"b", "1"},
		{"b", "2"},
		{"b", "3"},
	}, Arrangement([][]string{
		{"a", "b"},
		{"1", "2", "3"},
	}))

	assert.Equal(t, [][]string{
		{"a", "1", "x"},
		{"a", "1", "y"},
		{"a", "2", "x"},
		{"a", "2", "y"},
		{"a", "3", "x"},
		{"a", "3", "y"},
		{"b", "1", "x"},
		{"b", "1", "y"},
		{"b", "2", "x"},
		{"b", "2", "y"},
		{"b", "3", "x"},
		{"b", "3", "y"},
	}, Arrangement([][]string{
		{"a", "b"},
		{"1", "2", "3"},
		{"x", "y"},
	}))

	assert.Equal(t, [][]string{
		{"a", "1", "x"},
		{"a", "1", "y"},
		{"a", "2", "x"},
		{"a", "2", "y"},
		{"b", "1", "x"},
		{"b", "1", "y"},
		{"b", "2", "x"},
		{"b", "2", "y"},
	}, Arrangement([][]string{
		{"a", "b"},
		{"1", "2"},
		{"x", "y"},
	}))

	real := Arrangement([][]string{
		{"a", "b"},
		{"1", "2"},
		{"x", "y"},
		{"u", "v"},
	})
	assert.Equal(t, [][]string{
		{"a", "1", "x", "u"},
		{"a", "1", "x", "v"},
		{"a", "1", "y", "u"},
		{"a", "1", "y", "v"},
		{"a", "2", "x", "u"},
		{"a", "2", "x", "v"},
		{"a", "2", "y", "u"},
		{"a", "2", "y", "v"},
		{"b", "1", "x", "u"},
		{"b", "1", "x", "v"},
		{"b", "1", "y", "u"},
		{"b", "1", "y", "v"},
		{"b", "2", "x", "u"},
		{"b", "2", "x", "v"},
		{"b", "2", "y", "u"},
		{"b", "2", "y", "v"},
	}, real)
}

var n int

func f(s []int) {
	n = 0
	for _, v := range s {
		n += v
	}
}

func g(s []int) {
	x := 0
	for _, v := range s {
		x += v
	}
	n = x
}

var arr []int

func init() {
	arr = make([]int, 100000)
	for i := range arr {
		arr[i] = rand.Intn(100)
	}
}

func BenchmarkF(b *testing.B) {
	for i := 0; i < b.N; i++ {
		f(arr)
	}
}

func BenchmarkG(b *testing.B) {
	for i := 0; i < b.N; i++ {
		g(arr)
	}
}
