package database

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

type BTree struct {
	root *node
}

type node struct {
	keys         []string
	children     []*node
	childPageIDs []uint32
	keyLocations []*KeyLocation
	Next         *node
	NextID       uint32
	leaf         bool
	pageID       uint32
	pages        *Pages
}
type KeyLocation struct {
	PageID uint32
	SlotID uint16
}

func (Pages *Pages) NewBTree() (*BTree, uint32, error) {
	rootNode, err := Pages.newNode(true)
	Pages.rootPageLink(rootNode.pageID)
	if err != nil {
		return nil, rootNode.pageID, fmt.Errorf("failed to create new b+tree: %w", err)
	}
	t := &BTree{
		root: rootNode,
	}
	return t, rootNode.pageID, nil
}

func (Pages *Pages) newNode(isLeaf bool) (*node, error) {
	var pageID uint32
	var err error
	if isLeaf == true {
		pageID, err = Pages.Create(LeafPage)
		if err != nil {
			return nil, fmt.Errorf("failed to create new leaf node: %w", err)
		}
	} else {
		pageID, err = Pages.Create(LeafPage)
		if err != nil {
			return nil, fmt.Errorf("failed to create new leaf node: %w", err)
		}
	}
	n := &node{
		leaf:   isLeaf,
		pageID: pageID,
		pages:  Pages,
	}
	return n, err
}

func newKeyLocation(pageID uint32, slotID uint16) *KeyLocation {
	return &KeyLocation{
		PageID: pageID,
		SlotID: slotID,
	}
}

// A wrapper for insert to handle spliting of the root
func (DB *DB) Insert(key string, pageID uint32, slotID uint16) {
	KeyLocation := newKeyLocation(pageID, slotID)
	splitResult, left, right := DB.Pages.insert(DB.T.root, key, KeyLocation)
	if splitResult != nil {
		newRoot, _ := DB.Pages.newNode(false) // IMPORTANT I still need to catch my error
		newRoot.addKey(*splitResult, left, right, nil)
		DB.Pages.writeNodeToPage(newRoot)
		DB.T.root = newRoot
		DB.Pages.rootPageLink(newRoot.pageID)
	}
}

// Find the location of the key return KeyLocation and if found
func (DB *DB) FindKeyLocation(key string) (*KeyLocation, bool, error) {
	n, inTree, err := DB.T.findNode(key)
	if err != nil {
		return nil, false, fmt.Errorf("failed to find node: %w", err)
	}
	if !inTree {
		return nil, false, nil
	}
	i := slices.Index(n.keys, key)
	keyLocation := n.keyLocations[i]
	return keyLocation, true, nil
}

// Returns if key in tree and key
func (t *BTree) findNode(key string) (*node, bool, error) {
	var err error
	n := t.root
	for {
		if n.leaf {
			if slices.Contains(n.keys, key) {
				return n, true, nil
			} else {
				return n, false, nil
			}
		}
		n, _, err = n.findChild(key)
		if err != nil {
			return nil, false, fmt.Errorf("failed to find child: %w", err)
		}
	}
}

/*
Recursively calls itself to create a call stack,
adds key at leaf, and works back up splitting if needed
*/
func (Pages *Pages) insert(n *node, key string, KeyLocation *KeyLocation) (*string, *node, *node) {
	if n.leaf {
		splitResult, left, right := n.addKey(key, nil, nil, KeyLocation)
		if splitResult == nil {
			Pages.writeNodeToPage(n)
		}
		return splitResult, left, right
	}
	// Internal node
	child, _, _ := n.findChild(key)
	splitResult, left, right := Pages.insert(child, key, KeyLocation)

	if splitResult != nil {
		splitResult, left, right = n.addKey(*splitResult, left, right, nil)
		if splitResult == nil {
			Pages.writeNodeToPage(n)
		}
	} else {
		Pages.writeNodeToPage(n)
	}
	return splitResult, left, right
}

func (n *node) addKey(key string, left *node, right *node, KeyLocation *KeyLocation) (*string, *node, *node) {
	var splitResult *string
	// Checks if key in node already
	if n.leaf && slices.Contains(n.keys, key) {
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
	if n.leaf {
		n.keyLocations = append(n.keyLocations, nil)
		copy(n.keyLocations[i+1:], n.keyLocations[i:])
		n.keyLocations[i] = KeyLocation
	}

	n.addChildern(left, right)
	left = nil
	right = nil

	if n.isOverflow() {
		splitResult, left, right = n.split()
	}

	return splitResult, left, right
}

/*
Used when a node has to split as too many keys
So children also get split
*/
func (n *node) addChildern(left *node, right *node) {
	if left == nil || right == nil {
		return
	}
	if n.leaf {
		n.leaf = false
	}
	leftChildKey := left.keys[len(left.keys)-1]
	i := sort.Search(len(n.keys), func(i int) bool {
		return strings.Compare(leftChildKey, n.keys[i]) < 0
	})

	// Insert children left at odd child postion
	if i+1 > len(n.children) {
		n.children = append(n.children, nil)
		n.childPageIDs = append(n.childPageIDs, 0)
	}
	n.children[i] = left
	n.childPageIDs[i] = left.pageID

	n.children = append(n.children, nil)
	copy(n.children[i+2:], n.children[i+1:])
	n.children[i+1] = right

	n.childPageIDs = append(n.childPageIDs, 0)
	copy(n.childPageIDs[i+2:], n.childPageIDs[i+1:])
	n.childPageIDs[i+1] = right.pageID
}

// splits node into two new nodes
func (n *node) split() (*string, *node, *node) {
	right, _ := n.pages.newNode(n.leaf)

	middlekeyIndex := len(n.keys) / 2
	middlekey := &n.keys[middlekeyIndex]

	// Updates keys
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
	if n.leaf {
		rightStart := middlekeyIndex
		right.keyLocations = make([]*KeyLocation, len(n.keyLocations[rightStart:]))
		copy(right.keyLocations, n.keyLocations[rightStart:])
		n.keyLocations = n.keyLocations[:middlekeyIndex]
	}

	// Updates children (splits into different nodes)
	if len(n.children) > middlekeyIndex {
		rightChildStart := middlekeyIndex + 1

		right.children = append(right.children, n.children[rightChildStart:]...)
		n.children = n.children[:rightChildStart]

		right.childPageIDs = append(right.childPageIDs, n.childPageIDs[rightChildStart:]...)
		n.childPageIDs = n.childPageIDs[:rightChildStart]
	}

	// Updates next node
	right.Next = n.Next
	right.NextID = n.NextID

	n.Next = right
	n.NextID = right.pageID

	// Write node to disk
	n.pages.writeNodeToPage(n)
	n.pages.writeNodeToPage(right)

	return middlekey, n, right
}

// Finds the child where the key should be inesrt at. Read the child if not in memory
func (n *node) findChild(key string) (*node, int, error) {
	for i := range len(n.keys) {
		if strings.Compare(key, n.keys[i]) < 0 {

			if n.children[i] != nil {
				return n.children[i], i, nil
			}
			// Reads child from disk
			childNode, err := n.pages.ReadPageNode((n.childPageIDs[i]))
			n.children[i] = childNode
			if err != nil {
				return nil, i, fmt.Errorf("failed to get child at index %d: %w", i, err)
			}
			return childNode, i, nil
		}
	}
	// Last child
	if n.children[len(n.keys)] != nil {
		return n.children[len(n.keys)], len(n.keys), nil
	}
	// Reads child from disk
	childNode, err := n.pages.ReadPageNode((n.childPageIDs[len(n.keys)]))
	n.children[len(n.keys)] = childNode
	if err != nil {
		return nil, len(n.keys), fmt.Errorf("failed to get child at index %d: %w", len(n.keys), err)
	}
	return childNode, len(n.keys), nil
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

// Used for debugging
func (t *BTree) PrintLinkedList() {
	n := t.root
	for !n.leaf {
		n = n.children[0]
	}

	for n != nil {
		fmt.Println(n.keys)

		switch {
		case n.Next != nil:
			n = n.Next
		case n.NextID != 0:
			n, _ = n.pages.ReadPageNode(n.NextID)
		default:
			n = nil
		}
	}
}

// made this maybe for debugging
func PrintOutNode(n *node) {
	fmt.Println()
	fmt.Println("n.pageID:", n.pageID)
	fmt.Println("n.leaf:", n.leaf)
	fmt.Println("n.keys:", n.keys)

	if n.leaf == true {
		fmt.Println("n.Next:", n.Next)
		fmt.Println("n.NextID:", n.NextID)
		for _, keyLocation := range n.keyLocations {
			fmt.Println("pageID:", keyLocation.PageID)
			fmt.Println("slotID:", keyLocation.SlotID)
		}
	} else {
		fmt.Println("n.children:", n.children)
		fmt.Println("n.childPageIDs:", n.childPageIDs)
	}
}
