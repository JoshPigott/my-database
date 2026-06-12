package btrees

import (
	"slices"
)

// Handles making the tree shallower
func (t *BTree) Delete(key int) {
	t.root.delete(key)

	if len(t.root.keys) == 0 && len(t.root.children) > 0 {
		t.root = t.root.children[0]
	}
}

/*
Build a call stack to locate the leaf node.
Delete the key from the leaf node.
Rebalance the tree if necessary.
*/
func (n *node) delete(key int) {
	if n.leaf == true {
		if !slices.Contains(n.keys, key) {
			return
		}
		// Delete key and KeyLocation
		for i, keyVal := range n.keys {
			if keyVal == key {
				n.keys = append(n.keys[:i], n.keys[i+1:]...)
				n.KeyLocations = append(n.KeyLocations[:i], n.KeyLocations[i+1:]...)
				return
			}
		}
	}
	nChild, i := n.findChild(key)
	nChild.delete(key)

	childkeyNum := len(n.children[i].keys)
	underflow := childkeyNum < minKeys
	if underflow {
		// try borrow key from child sibbling
		isLeft := i != 0
		if isLeft {
			leftKeyLen := len(n.children[i-1].keys)
			if leftKeyLen > minKeys {
				n.borrowFromLeft(i)
				return
			}
		}
		isRight := i != len(n.children)-1
		if isRight {
			rightKeyLen := len(n.children[i+1].keys)
			if rightKeyLen > minKeys {
				n.borrowFromRight(i)
				return
			}
		}
		if isLeft {
			n.mergeWithLeft(i)
			return
		}
		n.mergeWithRight(i)
	}
}

// Borrow key right most key from left node to rebalance tree
func (n *node) borrowFromLeft(i int) {
	var separatorKey int
	left := n.children[i-1]
	borrowedKey := left.keys[len(left.keys)-1]
	borrowedKeyLocation := left.KeyLocations[len(left.KeyLocations)-1]
	if left.leaf == true {
		separatorKey = borrowedKey
	} else {
		separatorKey = n.children[i].children[0].keys[0]
	}

	// Update keys
	left.keys = left.keys[:len(left.keys)-1]
	n.keys[i-1] = borrowedKey
	n.children[i].keys = slices.Insert(n.children[i].keys, 0, separatorKey)

	// Update key locations
	if left.leaf == true {
		left.KeyLocations = left.KeyLocations[:len(left.KeyLocations)-1]
		n.children[i].KeyLocations = slices.Insert(n.children[i].KeyLocations, 0, borrowedKeyLocation)
	}

	// Move children
	if left.leaf == false {
		borrowedChild := left.children[len(left.children)-1]
		left.children = left.children[:len(left.children)-1]
		n.children[i].children = append([]*node{borrowedChild}, n.children[i].children...)
	}
}

// Borrow key left most key from right node to rebalance tree
func (n *node) borrowFromRight(i int) {
	right := n.children[i+1]
	borrowedKey := right.keys[0]
	separatorKey := right.keys[1]
	borrowedKeyLocation := right.KeyLocations[0]

	// Update keys
	right.keys = right.keys[1:]
	n.keys[i] = separatorKey
	n.children[i].keys = append(n.children[i].keys, borrowedKey)

	// Update key locations
	if right.leaf == true {
		right.KeyLocations = right.KeyLocations[1:]
		n.children[i].KeyLocations = append(n.children[i].KeyLocations, borrowedKeyLocation)
	}

	// Move children
	if right.leaf == false {
		borrowedChild := right.children[0]
		right.children = right.children[1:]
		n.children[i].children = append(n.children[i].children, borrowedChild)
	}
}

// Merges child node with their left sibling node
func (n *node) mergeWithLeft(i int) {
	left := n.children[i-1]

	// Delete separator key
	separatorKey := n.keys[i-1]
	for j, keyVal := range n.keys {
		if keyVal == separatorKey {
			n.keys = append(n.keys[:j], n.keys[j+1:]...)
			break
		}
	}

	// Updates keys and key location
	internalNode := n.children[i-1].leaf == false
	if internalNode {
		left.keys = append(left.keys, separatorKey)
		left.keys = append(left.keys, n.children[i].keys...)
	} else {
		left.keys = append(left.keys, n.children[i].keys...)
		left.KeyLocations = append(left.KeyLocations, n.children[i].KeyLocations...)
	}

	// Updates linked list
	left.Next = n.children[i].Next

	// Updates children
	left.children = append(left.children, n.children[i].children...)
	n.children = append(n.children[:i], n.children[i+1:]...)
}

// Merges child node with their right sibling node
func (n *node) mergeWithRight(i int) {
	right := n.children[i+1]

	// Delete separator key
	separatorKey := n.keys[i]
	for j, keyVal := range n.keys {
		if keyVal == separatorKey {
			n.keys = append(n.keys[:j], n.keys[j+1:]...)
			break
		}
	}

	// Updates keys and key location
	internalNode := n.children[i].leaf == false
	if internalNode {
		n.children[i].keys = append(n.children[i].keys, separatorKey)
		n.children[i].keys = append(n.children[i].keys, right.keys...)
	} else {
		n.children[i].keys = append(n.children[i].keys, right.keys...)
		n.children[i].KeyLocations = append(n.children[i].KeyLocations, right.KeyLocations...)
	}

	// Updates linked list
	n.children[i].Next = right.Next

	// Updates children
	n.children[i].children = append(n.children[i].children, right.children...)
	// Removes right
	n.children = append(n.children[:i+1], n.children[i+2:]...)
}
