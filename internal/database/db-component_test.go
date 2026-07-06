package database

import (
	"fmt"
	"testing"
)

// Low key jsut going to try and do the unit / componentm tests to start of with not i/o disk stuff

// Lets just do the fomrating and unformating pages tonight
// Maybe later try and add some edge cases on

type splitTestCase struct {
	name          string
	wantL         *node
	wantR         *node
	wantMiddleKey string
	inputL        *node
	inputR        *node
}

func Test_convertFormatInternalNode(t *testing.T) {
	numOfKey := 10
	n := genInternalNode(numOfKey, 0)
	wantN := n

	pages := Pages{
		pageIDToNode: map[uint32]*node{},
	}

	buf := createInternalNodeBuf(n)
	gotN := pages.formatNode(buf)

	helperCheckNode(t, gotN, wantN)
	helperCheckChildern(t, gotN, wantN)
}

func Test_convertFormatLeafNode(t *testing.T) {
	numOfKey := 10
	n := genLeafNode(numOfKey, 0)
	wantN := n

	pages := Pages{
		pageIDToNode: map[uint32]*node{},
	}

	buf := createLeafNodeBuf(n)
	gotN := pages.formatNode(buf)

	helperCheckNode(t, gotN, wantN)
	if gotN.NextID != wantN.NextID {
		t.Errorf("incorrect next page id: got %v, want %v", gotN.NextID, wantN.NextID)
	}
	helperCheckKeyLocations(t, gotN, wantN)
}

// Here I should do some test cases where I spit leaf and internal node I think
func Test_computeSplit(t *testing.T) {
	testcases := []splitTestCase{
		makeSplitTestCase("leaf even keys", true, 200),
		makeSplitTestCase("leaf odd keys", true, 199),
		makeSplitTestCase("internal even keys", false, 200),
		makeSplitTestCase("internal odd keys", false, 199),
	}
	// Make a for loop here
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			gotMiddleKey := computeSplit(tc.inputL, tc.inputR)
			// Note input are pointer so value got updated
			gotL := tc.inputL
			gotR := tc.inputR

			// Do some checking
			helperCheckNode(t, gotL, tc.wantL)
			helperCheckNode(t, gotR, tc.wantR)

			if *gotMiddleKey != tc.wantMiddleKey {
				t.Errorf("incorrect middle key for %s: got %s, want %s", tc.name, *gotMiddleKey, tc.wantMiddleKey)
			}
			if tc.wantL.leaf {
				if gotL.NextID != tc.wantL.NextID {
					t.Errorf("incorrect next page id for %s: got %v, want %v", tc.name, gotL.NextID, tc.wantL.NextID)
				}
				if gotR.NextID != tc.wantR.NextID {
					t.Errorf("incorrect next page id: got %v, want %v", gotR.NextID, tc.wantR.NextID)
				}
				helperCheckKeyLocations(t, gotL, tc.wantL)
				helperCheckKeyLocations(t, gotR, tc.wantR)
			} else {
				helperCheckChildern(t, gotL, tc.wantL)
				helperCheckChildern(t, gotR, tc.wantR)
			}
		})
	}
}

func makeSplitTestCase(name string, isLeaf bool, overflowingKeys int) splitTestCase {
	noKeys := 0
	rPageID := uint32(4)

	keysAfterL := overflowingKeys / 2
	lStart := 0
	keysAfterR := overflowingKeys - keysAfterL
	rStart := keysAfterL

	wantL := genNode(isLeaf, keysAfterL, lStart)

	// Leaf nodes have linked list
	if isLeaf {
		wantL.NextID = rPageID
	}
	wantR := genNode(isLeaf, keysAfterR, rStart)
	wantR.pageID = rPageID

	inputL := genNode(isLeaf, overflowingKeys, 0)
	inputR := genNode(isLeaf, noKeys, 0)
	inputR.pageID = uint32(4)

	wantMiddleKey := wantR.keys[0]
	// As seprator key does not stay in child nodes
	if !isLeaf {
		wantR.keys = wantR.keys[1:]
		wantR.childPageIDs = wantR.childPageIDs[1:]
		wantR.children = wantR.children[1:]
	}

	return splitTestCase{
		name:          name,
		wantL:         wantL,
		wantR:         wantR,
		wantMiddleKey: wantMiddleKey,
		inputL:        inputL,
		inputR:        inputR,
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
		t.Fatalf("incorrect key length: got %d, want %d", len(gotN.keys), len(wantN.keys))
	}
	for i := range wantN.keys {
		if gotN.keys[i] != wantN.keys[i] {
			t.Errorf("incorrect key at index %d: got %s, want %s", i, gotN.keys[i], wantN.keys[i])
		}
	}
}

func helperCheckKeyLocations(t *testing.T, gotN *node, wantN *node) {
	// Check the size against each other
	if len(gotN.keyLocations) != len(wantN.keyLocations) {
		t.Fatalf("incorrect key location length: got %d, want %d", len(gotN.keyLocations), len(wantN.keyLocations))
	}
	for i := range wantN.keyLocations {
		wantKeyLocations := wantN.keyLocations[i]
		wantPageID := wantKeyLocations.PageID
		wantSlotID := wantKeyLocations.SlotID

		gotKeyLocation := gotN.keyLocations[i]
		gotPageID := gotKeyLocation.PageID
		gotSlotID := gotKeyLocation.SlotID

		if gotPageID != wantPageID {
			t.Errorf("incorrect key location page id %d: got %d, want %d", i, gotPageID, wantPageID)
		}
		if gotSlotID != wantSlotID {
			t.Errorf("incorrect key location slot id %d: got %d, want %d", i, gotSlotID, wantSlotID)
		}
	}
}

func helperCheckChildern(t *testing.T, gotN *node, wantN *node) {
	if len(gotN.childPageIDs) != len(wantN.childPageIDs) {
		t.Fatalf("incorrect child page id length: got %d, want %d", len(gotN.childPageIDs), len(wantN.childPageIDs))
	}
	if len(gotN.children) != len(wantN.children) {
		t.Fatalf("incorrect key length: got %d, want %d", len(gotN.children), len(wantN.children))
	}
	for i := range wantN.childPageIDs {
		if gotN.childPageIDs[i] != wantN.childPageIDs[i] {
			t.Errorf("incorrect child page id at index %d: got %d, want %d", i, gotN.childPageIDs[i], wantN.childPageIDs[i])
		}

		if gotN.children[i] != wantN.children[i] {
			t.Errorf("incorrect child index %d: got %v, want %v", i, gotN.children[i], wantN.children[i])
		}
	}
}

// Generates a fake internal node
func genInternalNode(numKeys int, startNum int) *node {
	numChildren := numKeys + 1
	if numKeys == 0 {
		numChildren--
	}

	defaulPageID := uint32(6)
	// Children are lazy loaded
	children := make([]*node, numChildren)
	n := &node{
		keys:         genKeys(numKeys, startNum),
		childPageIDs: genChildPageIDs(numChildren, startNum),
		children:     children,
		pageID:       defaulPageID,
		leaf:         false,
	}
	return n
}

func genNode(leaf bool, numKeys int, startNum int) *node {
	if leaf {
		return genLeafNode(numKeys, startNum)
	} else {
		return genInternalNode(numKeys, startNum)
	}
}

// Generates a fake leaf node
func genLeafNode(numKeys int, startNum int) *node {
	// Undefined next page (no next page)
	defaultPageID := uint32(3)
	var nextPageID uint32 = 0

	n := &node{
		keys:         genKeys(numKeys, startNum),
		keyLocations: genKeylocation(numKeys, startNum),
		NextID:       nextPageID,
		pageID:       defaultPageID,
		leaf:         true,
	}
	return n
}

// Generates fake keys
func genKeys(numKeys int, startNum int) []string {
	keys := make([]string, numKeys)
	for i := range numKeys {
		keys[i] = fmt.Sprintf("entry%04d", i+startNum)
	}
	return keys
}

// Generates fake keys locations
func genKeylocation(numkeyLocations int, startNum int) []*KeyLocation {
	keyLocations := make([]*KeyLocation, numkeyLocations)

	for i := range numkeyLocations {
		keyLocation := KeyLocation{
			PageID: uint32(i + startNum),
			SlotID: uint16(i + startNum),
		}
		keyLocations[i] = &keyLocation
	}
	return keyLocations
}

// Generates fake child pages ids
func genChildPageIDs(numChildren int, startNum int) []uint32 {
	childPageIDs := make([]uint32, numChildren)
	for i := range numChildren {
		childPageIDs[i] = uint32(i + startNum)
	}
	return childPageIDs
}
