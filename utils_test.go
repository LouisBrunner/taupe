package taupe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMin(t *testing.T) {
	cases := []struct {
		input1 int
		input2 int
		output int
	}{
		{1, 2, 1},
		{1, -2, -2},
		{1, 1, 1},
	}
	for _, test := range cases {
		assert.Equal(t, test.output, imin(test.input1, test.input2))
	}
}

func TestMax(t *testing.T) {
	cases := []struct {
		input1 int
		input2 int
		output int
	}{
		{1, 2, 2},
		{1, -2, 1},
		{1, 1, 1},
	}
	for _, test := range cases {
		assert.Equal(t, test.output, imax(test.input1, test.input2))
	}
}

func TestLJust(t *testing.T) {
	cases := []struct {
		input1 string
		input2 int
		output string
	}{
		{"abc", 5, "abc  "},
		{"abc", 3, "abc"},
	}
	for _, test := range cases {
		assert.Equal(t, test.output, ljust(test.input1, test.input2))
	}

	assert.Panics(t, func() {
		ljust("abc", 2)
	})
}
