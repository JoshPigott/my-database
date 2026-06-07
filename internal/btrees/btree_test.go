package btrees

import (
	"testing"
)

func BenchmarkInerst(b *testing.B) {
	t := NewBTree()
	for i := 0; i < b.N; i++ {
		t.Insert(73)
	}
}
