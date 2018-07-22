package taupe

import (
	"strings"
)

func imin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func imax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ljust(s string, total int) string {
	return s + strings.Repeat(" ", total-len(s))
}
