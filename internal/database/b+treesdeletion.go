package database

import (
	"errors"
	"fmt"
	"slices"
)

// Handles making the tree shallower
func (t *BTree) Delete(key string) error {
	if err := t.root.delete(key); err != nil {
		return err
	}

	if len(t.root.keys) == 0 && len(t.root.children) > 0 {
		oldRoot := t.root
		t.root = t.root.children[0]
		t.root.pages.rootPageLink(t.root.pageID)
		oldRoot.pages.deleteNodePage(oldRoot)
	}
	return nil
}

/*
Build a call stack to locate the leaf node.
Delete the key from the leaf node.
Rebalance the tree if necessary.
*/
func (n *node) delete(key string) error {
	if n.leaf {
		if !slices.Contains(n.keys, key) {
			return nil
		}
		// Delete key and KeyLocation
		for i, keyVal := range n.keys {
			if keyVal == key {
				n.keys = append(n.keys[:i], n.keys[i+1:]...)
				n.KeyLocations = append(n.KeyLocations[:i], n.KeyLocations[i+1:]...)
				n.pages.writeNodeToPage(n)
				return nil
			}
		}
	}
	nChild, i, _ := n.findChild(key)
	if err := nChild.delete(key); err != nil {
		return err
	}
	childkeyNum := len(n.children[i].keys)
	underflow := childkeyNum < minKeys
	if !underflow {
		return nil
	}

	isLeft := i != 0
	isRight := i != len(n.children)-1

	if isLeft {
		if err := n.loadChildren(i - 1); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
		leftKeyLen := len(n.children[i-1].keys)
		if leftKeyLen > minKeys {
			n.borrowFromLeft(i)
			return nil
		}
	}
	if isRight {
		if err := n.loadChildren(i + 1); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
		rightKeyLen := len(n.children[i+1].keys)
		if rightKeyLen > minKeys {
			n.borrowFromRight(i)
			return nil
		}
	}
	if isLeft {
		n.mergeWithLeft(i)
		return nil
	}
	n.mergeWithRight(i)
	return nil
}

// Borrow key right most key from left node to rebalance tree
func (n *node) borrowFromLeft(i int) error {
	var newSeparatorKey string
	var borrowedKeyLocation *KeyLocation

	left := n.children[i-1]
	borrowedKey := left.keys[len(left.keys)-1]

	if left.leaf {
		newSeparatorKey = borrowedKey
		borrowedKeyLocation = left.KeyLocations[len(left.KeyLocations)-1]
	} else {
		newSeparatorKey = n.keys[i-1]
	}

	// Update keys
	left.keys = left.keys[:len(left.keys)-1]
	n.keys[i-1] = borrowedKey
	n.children[i].keys = slices.Insert(n.children[i].keys, 0, newSeparatorKey)

	// Update key locations
	if left.leaf {
		left.KeyLocations = left.KeyLocations[:len(left.KeyLocations)-1]
		n.children[i].KeyLocations = slices.Insert(n.children[i].KeyLocations, 0, borrowedKeyLocation)
	}

	// Move children
	if !left.leaf {
		borrowedChild := left.children[len(left.children)-1]
		left.children = left.children[:len(left.children)-1]
		n.children[i].children = append([]*node{borrowedChild}, n.children[i].children...)

		left.childPageIDs = left.childPageIDs[:len(left.childPageIDs)-1]
		n.children[i].childPageIDs = append([]uint32{borrowedChild.pageID}, n.children[i].childPageIDs...)
	}
	// Writes nodes to disk
	n.pages.writeNodeToPage(n)
	n.children[i].pages.writeNodeToPage(n.children[i])
	left.pages.writeNodeToPage(left)
	return nil
}

// Borrow key left most key from right node to rebalance tree
func (n *node) borrowFromRight(i int) {
	right := n.children[i+1]
	if right.leaf {
		borrowedKey := right.keys[0]
		right.keys = right.keys[1:]
		n.children[i].keys = append(n.children[i].keys, borrowedKey)
		n.keys[i] = right.keys[0]
	} else {
		borrowedKey := right.keys[0]
		right.keys = right.keys[1:]
		n.children[i].keys = append(n.children[i].keys, n.keys[i])
		n.keys[i] = borrowedKey
	}
	// Update key locations
	if right.leaf {
		borrowedKeyLocation := right.KeyLocations[0]
		right.KeyLocations = right.KeyLocations[1:]
		n.children[i].KeyLocations = append(n.children[i].KeyLocations, borrowedKeyLocation)
	}

	// Move children
	if !right.leaf {
		borrowedChild := right.children[0]
		right.children = right.children[1:]
		n.children[i].children = append(n.children[i].children, borrowedChild)

		right.childPageIDs = right.childPageIDs[1:]
		n.children[i].childPageIDs = append(n.children[i].childPageIDs, borrowedChild.pageID)
	}
	// Writes nodes to disk
	n.pages.writeNodeToPage(n)
	n.children[i].pages.writeNodeToPage(n.children[i])
	right.pages.writeNodeToPage(right)
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
	if left.leaf {
		left.Next = n.children[i].Next
		left.NextID = n.children[i].NextID
	}

	// Updates children
	merged := n.children[i]
	left.children = append(left.children, n.children[i].children...)
	n.children = append(n.children[:i], n.children[i+1:]...)

	left.childPageIDs = append(left.childPageIDs, merged.childPageIDs...)
	n.childPageIDs = append(n.childPageIDs[:i], n.childPageIDs[i+1:]...)

	// Write nodes to disk
	n.pages.writeNodeToPage(n)
	left.pages.writeNodeToPage(left)
	merged.pages.deleteNodePage(merged)
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
	if n.children[i].leaf {
		n.children[i].Next = right.Next
		n.children[i].NextID = right.NextID
	}

	// Updates children
	n.children[i].children = append(n.children[i].children, right.children...)
	n.children[i].childPageIDs = append(n.children[i].childPageIDs, right.childPageIDs...)
	// Removes right
	n.children = append(n.children[:i+1], n.children[i+2:]...)
	n.childPageIDs = append(n.childPageIDs[:i+1], n.childPageIDs[i+2:]...)

	// Write nodes to disk
	n.pages.writeNodeToPage(n)
	n.children[i].pages.writeNodeToPage(n.children[i])
	right.pages.deleteNodePage(right)
}

// Make sure right or left child is load into memory
func (n *node) loadChildren(j int) error {
	var err error
	if n.children[j] != nil {
		return nil
	}
	if n.childPageIDs[j] == 0 {
		return errors.New("failed to load left child")
	}
	n.children[j], err = n.pages.ReadPageNode(n.childPageIDs[j])
	if err != nil {
		return fmt.Errorf("failed to read right child: %w", err)
	}
	return nil
}
