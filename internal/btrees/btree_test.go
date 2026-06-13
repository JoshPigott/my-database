package btrees

import (
	"testing"
)

func BenchmarkInerst(b *testing.B) {
	t := NewBTree()
	for i := 0; i < b.N; i++ {
		keyLocation := KeyLocation{
			SlotID: 9,
			PageID: 3,
		}
		t.Insert("A", &keyLocation)
	}
}
