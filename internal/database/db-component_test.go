package database

import (
	"fmt"
	"slices"
	"testing"
)

type nodeParams struct {
	isLeaf     bool
	numKeys    int
	startIndex int
	keyStep    int
	pageID     uint32
}

type splitTestCase struct {
	name          string
	wantL         *node
	wantR         *node
	wantMiddleKey string
	inputL        *node
	inputR        *node
}

type addChildrenTestCase struct {
	name   string
	inputN *node
	inputL *node
	inputR *node
	i      int
	wantN  *node
}

func TestComponent_convertFormatInternalNode(t *testing.T) {
	n := genInternalNode(nodeParams{isLeaf: false, numKeys: 10, startIndex: 0, keyStep: 1})
	wantN := n.cloneNode()

	pages := Pages{
		pageIDToNode: map[uint32]*node{},
	}

	buf := createInternalNodeBuf(n)
	gotN := pages.formatNode(buf)

	helperCheckNode(t, gotN, wantN)
	helperCheckChildern(t, gotN, wantN)
}

func TestComponent_convertFormatLeafNode(t *testing.T) {
	n := genLeafNode(nodeParams{isLeaf: true, numKeys: 10, startIndex: 0, keyStep: 1})
	wantN := n.cloneNode()

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

func TestUnit_computeSplit(t *testing.T) {
	tcs := []splitTestCase{
		makeSplitTestCase("leaf even keys", true, 200),
		makeSplitTestCase("leaf odd keys", true, 199),
		makeSplitTestCase("internal even keys", false, 200),
		makeSplitTestCase("internal odd keys", false, 199),
	}
	// Make a for loop here
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			gotMiddleKey := computeSplit(tc.inputL, tc.inputR)
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

func TestUnit_addChildern(t *testing.T) {
	tcs := makeAddChildrenTestCases()
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.inputN.addChildern(tc.inputL, tc.inputR, tc.i)
			gotN := tc.inputN
			helperCheckNode(t, gotN, tc.wantN)
			helperCheckChildern(t, gotN, tc.wantN)
		})
	}
}

// Creates test cases of add children into a parent and start, middle and end
func makeAddChildrenTestCases() []addChildrenTestCase {
	tcs := []addChildrenTestCase{}

	start := 20
	step := 20
	lPageID := uint32(2)
	rPageID := uint32(3)

	names := []string{"children at start", "children at middle", "children at end"}
	iValues := []int{0, 5, 9}

	for j, name := range names {
		i := iValues[j]
		lStart := (i-1)*step + start
		rStart := i*step + start

		n := genInternalNode(nodeParams{numKeys: 11, startIndex: start, keyStep: 20, pageID: 1})
		l := genInternalNode(nodeParams{numKeys: 10, startIndex: lStart, keyStep: 1, pageID: lPageID})
		r := genInternalNode(nodeParams{numKeys: 10, startIndex: rStart, keyStep: 1, pageID: rPageID})

		// Setup leaf child
		n.childPageIDs[i] = l.pageID
		n.children[i] = l

		// Add right child for want
		wantN := n.cloneNode()
		wantN.childPageIDs[i] = lPageID
		wantN.childPageIDs[i+1] = rPageID
		wantN.children[i] = l
		wantN.children[i+1] = r

		// Remove right child for input
		n.childPageIDs = slices.Delete(n.childPageIDs, i+1, i+2)
		n.children = slices.Delete(n.children, i+1, i+2)

		tc := addChildrenTestCase{
			name:   name,
			inputN: n,
			inputL: l,
			inputR: r,
			i:      i,
			wantN:  wantN,
		}
		tcs = append(tcs, tc)
	}
	emptyParentTC := makeAddChildrenEmptyParent()
	tcs = append(tcs, emptyParentTC)
	return tcs
}

// Cretes the test cases when two children get into a new root node
func makeAddChildrenEmptyParent() addChildrenTestCase {
	// An empty parent
	n := genInternalNode(nodeParams{numKeys: 1, startIndex: 20})
	l := genInternalNode(nodeParams{numKeys: 10, startIndex: 0, keyStep: 1, pageID: 2})
	r := genInternalNode(nodeParams{numKeys: 10, startIndex: 20, keyStep: 1, pageID: 3})
	n.childPageIDs = []uint32{}
	n.children = []*node{}

	// Want
	wantN := n.cloneNode()
	wantN.childPageIDs = []uint32{2, 3}
	wantN.children = []*node{l, r}

	return addChildrenTestCase{
		name:   "empty parent",
		inputN: n,
		inputL: l,
		inputR: r,
		i:      0,
		wantN:  wantN,
	}
}

// I don't like this isLeaf thing happening
func makeSplitTestCase(name string, isLeaf bool, overflowingKeys int) splitTestCase {
	rPageID := uint32(4)

	keysAfterL := overflowingKeys / 2
	keysAfterR := overflowingKeys - keysAfterL
	rStart := keysAfterL

	wantL := genNode(nodeParams{isLeaf: isLeaf, numKeys: keysAfterL, startIndex: 0, keyStep: 1})

	// Leaf nodes have linked list
	if isLeaf {
		wantL.NextID = rPageID
	}
	wantR := genNode(nodeParams{isLeaf: isLeaf, numKeys: keysAfterR, startIndex: rStart, keyStep: 1})
	wantR.pageID = rPageID

	inputL := genNode(nodeParams{isLeaf: isLeaf, numKeys: overflowingKeys, startIndex: 0, keyStep: 1})
	inputR := genNode(nodeParams{isLeaf: isLeaf, numKeys: 0, startIndex: 0, keyStep: 1})
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
	t.Helper()
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
	t.Helper()
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
	t.Helper()
	//t.Helper()
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

// Creates new memory addresses so node is copied
func (n *node) cloneNode() *node {
	c := *n
	c.keys = slices.Clone(n.keys)
	if c.leaf {
		for i, keyLocation := range n.keyLocations {
			if keyLocation == nil {
				continue
			}
			copiedKeyLocation := *keyLocation
			c.keyLocations[i] = &copiedKeyLocation
		}
	} else {
		c.childPageIDs = slices.Clone(n.childPageIDs)
		copiedSlice := n.children
		c.children = copiedSlice
		for i, child := range n.children {
			if child == nil {
				continue
			}
			copiedChild := *child
			c.children[i] = &copiedChild
		}
	}
	return &c
}

// Generates a fake internal node
func genInternalNode(np nodeParams) *node {
	defaulPageID := uint32(6)
	numChildren := np.numKeys + 1
	if np.numKeys == 0 {
		numChildren--
	}

	if np.pageID == uint32(0) {
		np.pageID = defaulPageID
	}

	// Children are lazy loaded
	children := make([]*node, numChildren)
	n := &node{
		keys:         genKeys(np.numKeys, np.startIndex, np.keyStep),
		childPageIDs: genChildPageIDs(numChildren, np.startIndex),
		children:     children,
		pageID:       np.pageID,
		leaf:         false,
	}
	return n
}

// Note if I keep add another argument should maybe make a struct
func genNode(np nodeParams) *node {
	if np.isLeaf {
		return genLeafNode(np)
	} else {
		return genInternalNode(np)
	}
}

// Generates a fake leaf node
func genLeafNode(np nodeParams) *node {
	// Undefined next page (no next page)
	defaultPageID := uint32(3)
	var nextPageID uint32 = 0

	n := &node{
		keys:         genKeys(np.numKeys, np.startIndex, np.keyStep),
		keyLocations: genKeylocation(np.numKeys, np.startIndex),
		NextID:       nextPageID,
		pageID:       defaultPageID,
		leaf:         true,
	}
	return n
}

// Generates fake keys
func genKeys(numKeys int, startNum int, step int) []string {
	keys := make([]string, numKeys)
	for i := range numKeys {
		value := i*step + startNum
		keys[i] = fmt.Sprintf("entry%04d", value)
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
