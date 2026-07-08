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

func (pages *Pages) newNode(isLeaf bool) (*node, error) {
	var pageID uint32
	var err error
	if isLeaf {
		pageID, err = pages.create(LeafPage)
		if err != nil {
			return nil, fmt.Errorf("failed to create new page for leaf node: %w", err)
		}
	} else {
		pageID, err = pages.create(InternalPage)
		if err != nil {
			return nil, fmt.Errorf("failed to create new page for interal node: %w", err)
		}
	}
	n := &node{
		leaf:   isLeaf,
		pageID: pageID,
		pages:  pages,
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
func (db *DB) Insert(key string, pageID uint32, slotID uint16) error {
	KeyLocation := newKeyLocation(pageID, slotID)
	splitResult, left, right, err := db.Root.insert(key, KeyLocation)
	if err != nil {
		return err
	}
	if splitResult != nil {
		newRoot, err := db.Pages.newNode(false)
		if err != nil {
			return fmt.Errorf("failed to create new root node: %w", err)
		}
		if _, _, _, err := newRoot.addKey(*splitResult, left, right, nil); err != nil {
			return fmt.Errorf("failed to add key to new root node: %w", err)
		}
		if err := newRoot.writeNodeToPage(); err != nil {
			return fmt.Errorf("failed to add new key page: %w", err)
		}
		db.Root = newRoot
		if err := newRoot.rootPageLink(); err != nil {
			return fmt.Errorf("failed to add new page link: %w", err)
		}
	}
	return nil
}

// Find the location of the key return KeyLocation and if found
func (db *DB) FindKeyLocation(key string) (*KeyLocation, bool, error) {
	n, inTree, err := db.Root.findNode(key)
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
func (r *node) findNode(key string) (*node, bool, error) {
	var err error
	n := r
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
func (n *node) insert(key string, keyLocation *KeyLocation) (*string, *node, *node, error) {
	if n.leaf {
		splitResult, left, right, err := n.addKey(key, nil, nil, keyLocation)
		if err != nil {

		}
		if splitResult == nil {
			if err := n.writeNodeToPage(); err != nil {
				return nil, nil, nil, fmt.Errorf("failed to write new new to page: %w", err)
			}
		}
		return splitResult, left, right, nil
	}
	// Internal node
	child, _, _ := n.findChild(key)
	splitResult, left, right, err := child.insert(key, keyLocation)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to insert key: %w", err)
	}
	return n.performSplit(splitResult, left, right)
}

// Perform node split into two nodes if need
func (n *node) performSplit(splitResult *string, left *node, right *node) (*string, *node, *node, error) {
	if splitResult == nil {
		if err := n.writeNodeToPage(); err != nil {
			return nil, nil, nil, fmt.Errorf("failed write node to disk: %w", err)
		}
		return nil, nil, nil, nil
	}
	splitResult, left, right, err := n.addKey(*splitResult, left, right, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	if splitResult == nil {
		if err := n.writeNodeToPage(); err != nil {
			return nil, nil, nil, fmt.Errorf("failed write node to disk: %w", err)
		}
	}
	return splitResult, left, right, nil
}

func (n *node) addKey(key string, left *node, right *node, keyLocation *KeyLocation) (*string, *node, *node, error) {
	var splitResult *string
	var err error
	// Checks if key in node already
	if n.leaf && slices.Contains(n.keys, key) {
		return nil, nil, nil, nil
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
		n.keyLocations[i] = keyLocation
	}

	if left != nil && right != nil {
		n.addChildern(left, right, i)
	}

	if n.isOverflow() {
		splitResult, left, right, err = n.split()
		if err != nil {
			return nil, nil, nil, err
		}
	}

	return splitResult, left, right, nil
}

// Adds left and right node into the parent children after a split where old child was
func (n *node) addChildern(left *node, right *node, i int) {
	// Parent is a new node
	if len(n.children) == 0 {
		n.children = append(n.children, nil)
		n.childPageIDs = append(n.childPageIDs, 0)

		n.children[i] = left
		n.childPageIDs[i] = left.pageID
	}

	// Adds new right child
	n.children = append(n.children, nil)
	copy(n.children[i+2:], n.children[i+1:])
	n.children[i+1] = right

	n.childPageIDs = append(n.childPageIDs, 0)
	copy(n.childPageIDs[i+2:], n.childPageIDs[i+1:])
	n.childPageIDs[i+1] = right.pageID

	left = nil
	right = nil
}

// splits node into two new nodes
func (n *node) split() (*string, *node, *node, error) {
	right, err := n.pages.newNode(n.leaf)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create new node: %w", err)
	}
	middlekey := computeSplit(n, right)
	// Write node to disk
	if err := n.writeNodeToPage(); err != nil {
		return nil, nil, nil, fmt.Errorf("failed write left spilt node to disk: %w", err)
	}
	if err := right.writeNodeToPage(); err != nil {
		return nil, nil, nil, fmt.Errorf("failed write right spilt node to disk: %w", err)
	}
	return middlekey, n, right, nil
}

func computeSplit(l *node, r *node) *string {
	middleIndex := len(l.keys) / 2
	middlekey := &l.keys[middleIndex]

	rStart := middleIndex
	if !l.leaf {
		rStart = middleIndex + 1
	}

	// Updates keys
	if !l.leaf {
		r.keys = make([]string, len(l.keys[rStart:]))
		copy(r.keys, l.keys[rStart:])
	} else {
		r.keys = make([]string, len(l.keys[rStart:]))
		copy(r.keys, l.keys[rStart:])
	}
	l.keys = l.keys[:middleIndex]

	// Updates KeyLocation
	if l.leaf {
		r.keyLocations = make([]*KeyLocation, len(l.keyLocations[rStart:]))
		copy(r.keyLocations, l.keyLocations[rStart:])
		l.keyLocations = l.keyLocations[:middleIndex]
	}

	// Updates children (splits into different nodes)
	if len(l.childPageIDs) > middleIndex {
		rChildStart := middleIndex + 1

		r.children = append(r.children, l.children[rChildStart:]...)
		l.children = l.children[:rChildStart]

		r.childPageIDs = append(r.childPageIDs, l.childPageIDs[rChildStart:]...)
		l.childPageIDs = l.childPageIDs[:rChildStart]
	}

	// Updates next node
	if l.leaf {
		r.Next = l.Next
		r.NextID = l.NextID

		l.Next = r
		l.NextID = r.pageID
	}
	return middlekey
}

// Finds the child where the key should be inesrt at. Read the child if not in memory
func (n *node) findChild(targetKey string) (*node, int, error) {
	for i, key := range n.keys {
		if strings.Compare(targetKey, key) < 0 {
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

// Used for debugging
func (r *node) PrintLinkedList() {
	n := r
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
