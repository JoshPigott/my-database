package btrees

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

const (
	m       = 5
	maxKeys = m - 1
	minKeys = m / 2
)

type BTree struct {
	root *node
}

type node struct {
	keys         []string
	children     []*node
	KeyLocations []*KeyLocation
	Next         *node
	leaf         bool
}

type KeyLocation struct {
	PageID uint32
	SlotID uint16
}

func NewBTree() *BTree {
	rootNode := newNode(true)
	return &BTree{
		root: rootNode,
	}
}

func newNode(isLeaf bool) *node {
	return &node{
		leaf: isLeaf,
	}
}

func newKeyLocation(pageID uint32, slotID uint16) *KeyLocation {
	return &KeyLocation{
		PageID: pageID,
		SlotID: slotID,
	}
}

// A wrapper for insert to handle spliting of the root
func (t *BTree) Insert(key string, pageID uint32, slotID uint16) {
	KeyLocation := newKeyLocation(pageID, slotID)
	splitResult, left, right := insert(t.root, key, KeyLocation)
	if splitResult != nil {
		newRoot := newNode(false)
		newRoot.addKey(*splitResult, left, right, nil)
		t.root = newRoot
	}
}

// Find the location of the key return KeyLocation and if found
func (t *BTree) FindKeyLocation(key string) (uint32, uint16, bool) {
	n, inTree := t.FindNode(key)
	if !inTree {
		return 0, 0, false
	}
	// I need to find the index where the key happens
	i := slices.Index(n.keys, key)
	keyLocation := *n.KeyLocations[i]
	return keyLocation.PageID, keyLocation.SlotID, true
}

// Returns if key in tree and key
func (t *BTree) FindNode(key string) (*node, bool) {
	n := t.root
	for {
		if n.leaf {
			if slices.Contains(n.keys, key) {
				return n, true
			} else {
				return nil, false
			}
		}
		n, _ = n.findChild(key)
	}
}

/*
Recursively calls itself to create a call stack,
adds key at leaf, and works back up splitting if needed
*/
func insert(n *node, key string, KeyLocation *KeyLocation) (*string, *node, *node) {
	if n.leaf {
		splitResult, left, right := n.addKey(key, nil, nil, KeyLocation)
		return splitResult, left, right
	}

	child, _ := n.findChild(key)
	splitResult, left, right := insert(child, key, KeyLocation)

	if splitResult != nil {
		splitResult, left, right = n.addKey(*splitResult, left, right, nil)
	}
	return splitResult, left, right
}

func (n *node) addKey(key string, left *node, right *node, KeyLocation *KeyLocation) (*string, *node, *node) {
	var splitResult *string
	// Checks if key in node already
	if n.leaf == true && slices.Contains(n.keys, key) {
		return nil, nil, nil
	}

	i := sort.Search(len(n.keys), func(i int) bool {
		return strings.Compare(key, n.keys[i]) < 0
	})

	// Add key
	n.keys = append(n.keys, "")
	copy(n.keys[i+1:], n.keys[i:])
	n.keys[i] = key

	// Add key location
	if n.leaf == true {
		n.KeyLocations = append(n.KeyLocations, nil)
		copy(n.KeyLocations[i+1:], n.KeyLocations[i:])
		n.KeyLocations[i] = KeyLocation
	}

	n.addChildern(left, right)
	left = nil
	right = nil

	if len(n.keys) > maxKeys {
		splitResult, left, right = n.split()
	}
	return splitResult, left, right
}

// Get insert postion and insert children left at i and right at i + 1
func (n *node) addChildern(left *node, right *node) {
	if left == nil || right == nil {
		return
	}
	if n.leaf == true {
		n.leaf = false
	}
	leftChildKey := left.keys[len(left.keys)-1]
	i := sort.Search(len(n.keys), func(i int) bool {
		return strings.Compare(leftChildKey, n.keys[i]) < 0
	})

	// Insert children
	if i+1 > len(n.children) {
		n.children = append(n.children, nil)
	}
	n.children[i] = left

	n.children = append(n.children, nil)
	copy(n.children[i+2:], n.children[i+1:])
	n.children[i+1] = right
}

// splits node into two new nodes
func (n *node) split() (*string, *node, *node) {
	right := newNode(true)

	middlekeyIndex := maxKeys / 2
	middlekey := &n.keys[middlekeyIndex]

	// Updates keys
	// If leaf spilt right key stays in leaf
	if n.leaf == false {
		rightStart := middlekeyIndex + 1
		right.keys = make([]string, len(n.keys[rightStart:]))
		copy(right.keys, n.keys[rightStart:])
	} else {
		rightStart := middlekeyIndex
		right.keys = make([]string, len(n.keys[rightStart:]))
		copy(right.keys, n.keys[rightStart:])
	}
	n.keys = n.keys[:middlekeyIndex]

	// Updates KeyLocation
	if n.leaf == true {
		rightStart := middlekeyIndex
		right.KeyLocations = make([]*KeyLocation, len(n.KeyLocations[rightStart:]))
		copy(right.KeyLocations, n.KeyLocations[rightStart:])
		n.KeyLocations = n.KeyLocations[:middlekeyIndex]
	}

	// Updates children
	if len(n.children) > middlekeyIndex {
		rightChildStart := middlekeyIndex + 1
		right.children = append(right.children, n.children[rightChildStart:]...)
		n.children = n.children[:rightChildStart]
		if len(right.children) > 0 {
			right.leaf = false
		}
		if len(n.children) > 0 {
			n.leaf = false
		}
	}

	// Updates next node
	right.Next = n.Next
	n.Next = right

	return middlekey, n, right
}

// Find next child of current node and it index
func (n *node) findChild(key string) (*node, int) {
	for i := range len(n.keys) {
		if strings.Compare(key, n.keys[i]) < 0 {
			return n.children[i], i
		}
	}
	return n.children[len(n.keys)], len(n.keys)
}

// Prints out the btree structure for debugging purposes
func (t *BTree) CheckStructure(number int) {
	fmt.Println("Number:", number)
	fmt.Println("t.root.keys:", t.root.keys)
	for i := range len(t.root.children) {
		stringCode := fmt.Sprintf("t.root.children[%d].keys:", i)
		fmt.Println(stringCode, t.root.children[i].keys)
		for j := range len(t.root.children[i].children) {
			stringCode := fmt.Sprintf("t.root.children[%d].children[%d].keys:", i, j)
			fmt.Println(stringCode, t.root.children[i].children[j].keys)
			for k := range len(t.root.children[i].children[j].children) {
				stringCode := fmt.Sprintf("t.root.children[%d].children[%d].children[%d].keys:", i, j, k)
				fmt.Println(stringCode, t.root.children[i].children[j].children[k].keys)
			}
		}
	}
	println()
}

// Used for debuging
func (t *BTree) PrintLinkedList() {
	node := t.root
	for node.leaf != true {
		node = node.children[0]
	}
	for node.Next != nil {
		fmt.Println(node.keys)
		node = node.Next
	}
	fmt.Println(node.keys)
}
