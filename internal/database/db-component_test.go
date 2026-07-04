package database

import "testing"

// Low key jsut going to try and do the unit / componentm tests to start of with not i/o disk stuff

// Lets just do the fomrating and unformating pages tonight
// Maybe later try and add some edge cases on

func Test_convertFormatInternalNode(t *testing.T) {
	// Fake internal node
	var pageID uint32 = 6
	isLeaf := false
	keys := []string{"entry0", "entry1", "entry2", "entry3",
		"entry4", "entry5", "entry6", "entry7", "entry8", "entry9"}
	childPageIDs := []uint32{7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}
	// Children are lazy loaded
	wantChildren := make([]*node, 11)

	pages := Pages{
		pageIDToNode: map[uint32]*node{},
	}

	n := &node{
		keys:         keys,
		NextID:       0,
		childPageIDs: childPageIDs,
		pageID:       pageID,
		leaf:         isLeaf,
	}

	buf := createInternalNodeBuf(n)
	gotN := pages.formatNode(buf)

	helperCheckNode(t, gotN, n)

	if len(gotN.childPageIDs) != len(childPageIDs) {
		t.Fatalf("incorrect key lengths: got %d, want %d", len(gotN.childPageIDs), len(childPageIDs))
	}
	if len(gotN.childPageIDs) != len(childPageIDs) {
		t.Fatalf("incorrect key lengths: got %d, want %d", len(gotN.childPageIDs), len(childPageIDs))
	}
	for i := range childPageIDs {
		if gotN.childPageIDs[i] != childPageIDs[i] {
			t.Errorf("incorrect child page id at index %d: got %d, want %d", i, gotN.childPageIDs[i], childPageIDs[i])
		}

		if gotN.children[i] != wantChildren[i] {
			t.Errorf("incorrect child index %d: got %v, want %v", i, gotN.children[i], wantChildren[i])
		}
	}
}

func Test_convertFormatLeafNode(t *testing.T) {
	// Fake leaf node
	var pageID uint32 = 3
	// Undefined next page (no next page)
	var nextPageID uint32 = 0
	isLeaf := true

	keys := []string{"entry0", "entry1", "entry2", "entry3",
		"entry4", "entry5", "entry6", "entry7", "entry8", "entry9"}
	pageIDs := []uint32{2, 2, 2, 2, 2, 2, 2, 4, 4, 4}
	slotIDs := []uint16{20, 21, 22, 23, 24, 25, 26, 0, 1, 2}
	keyLocations := make([]*KeyLocation, len(pageIDs))

	for i := range pageIDs {
		keyLocation := KeyLocation{
			PageID: pageIDs[i],
			SlotID: slotIDs[i],
		}
		keyLocations[i] = &keyLocation
	}

	pages := Pages{
		pageIDToNode: map[uint32]*node{},
	}

	n := &node{
		keys:         keys,
		keyLocations: keyLocations,
		NextID:       nextPageID,
		pageID:       pageID,
		leaf:         isLeaf,
	}

	buf := createLeafNodeBuf(n)
	gotN := pages.formatNode(buf)

	helperCheckNode(t, gotN, n)
	if gotN.NextID != nextPageID {
		t.Errorf("incorrect next page id: got %v, want %v", gotN.NextID, nextPageID)
	}

	for i := range pageIDs {
		gotKeyLocation := gotN.keyLocations[i]
		gotPageID := gotKeyLocation.PageID
		gotSlotID := gotKeyLocation.SlotID
		if gotPageID != pageIDs[i] {
			t.Errorf("incorrect key location page id %d: got %d, want %d", i, gotPageID, pageIDs[i])
		}
		if gotSlotID != slotIDs[i] {
			t.Errorf("incorrect key location slot id %d: got %d, want %d", i, gotSlotID, slotIDs[i])
		}
	}
}

// Checks shared node attributes; got vs want
func helperCheckNode(t *testing.T, gotN *node, wantN *node) {
	if gotN.pageID != wantN.pageID {
		t.Errorf("incorrect pageID: got %d, want %d", gotN.pageID, wantN.pageID)
	}
	if gotN.leaf != wantN.leaf {
		t.Errorf("incorrect leaf status: got %v, want %v", gotN.leaf, wantN.leaf)
	}
	if len(gotN.keys) != len(wantN.keys) {
		t.Fatalf("incorrect key lengths: got %d, want %d", len(gotN.keys), len(wantN.keys))
	}
	for i := range wantN.keys {
		if gotN.keys[i] != wantN.keys[i] {
			t.Errorf("incorrect key at index %d: got %s, want %s", i, gotN.keys[i], wantN.keys[i])
		}
	}
}
