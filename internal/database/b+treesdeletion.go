package database

import (
	"errors"
	"fmt"
	"math"
	"slices"
)

// Handles making the tree shallower
func (r *node) Delete(key string) error {
	if err := r.delete(key); err != nil {
		return err
	}

	if len(r.keys) == 0 && len(r.children) > 0 {
		oldRoot := r
		r = r.children[0]
		if err := r.rootPageLink(); err != nil {
			return fmt.Errorf("failed to update root page link: %w", err)
		}
		if err := oldRoot.deleteNodePage(); err != nil {
			return fmt.Errorf("failed to delete old root page: %w", err)
		}
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
				n.keyLocations = append(n.keyLocations[:i], n.keyLocations[i+1:]...)
				if err := n.writeNodeToPage(); err != nil {
					return fmt.Errorf("failed to write node to page: %w", err)
				}
				return nil
			}
		}
	}
	nChild, i, err := n.findChild(key)
	if err != nil {
		return fmt.Errorf("failed to find child: %w", err)
	}

	if err := nChild.delete(key); err != nil {
		return err
	}
	if nChild.isUnderflow() {
		return n.fixUnderflow(i)
	}
	return nil
}

// Updates child keys to ensure the child node has sufficient keys
func (n *node) fixUnderflow(i int) error {
	hasLeftSibling := i != 0
	hasRightSibling := i != len(n.children)-1

	// Load children pages
	if hasLeftSibling {
		if err := n.loadChildren(i - 1); err != nil {
			return fmt.Errorf("failed to load left child node: %w", err)
		}
	}
	if hasRightSibling {
		if err := n.loadChildren(i + 1); err != nil {
			return fmt.Errorf("failed to load rigth child node: %w", err)
		}
	}

	if hasLeftSibling && n.needsLeftRedistribution(i) {
		if err := n.leftRedistribution(i); err != nil {
			return fmt.Errorf("failed to left redistribution: %w", err)
		}
	} else if hasRightSibling && n.needsRightRedistribution(i) {
		if err := n.rightRedistribution(i); err != nil {
			return fmt.Errorf("failed to right redistribution: %w", err)
		}
	} else if hasLeftSibling {
		if err := n.mergeWithLeft(i); err != nil {
			return fmt.Errorf("failed to merge with left: %w", err)
		}
	} else if hasRightSibling {
		if err := n.mergeWithRight(i); err != nil {
			return fmt.Errorf("failed to merge with right: %w", err)
		}
	}
	return nil
}

func (n *node) leftRedistribution(i int) error {
	withLeft := true
	l, r := n.computeRedistribution(i, withLeft)
	return n.writeRedistribution(l, r)
}

func (n *node) rightRedistribution(i int) error {
	withLeft := false
	l, r := n.computeRedistribution(i, withLeft)
	return n.writeRedistribution(l, r)
}

// Rewrites child and right child to have the a more even split of data
func (n *node) computeRedistribution(i int, withLeft bool) (*node, *node) { // I should test this
	var l *node
	var r *node
	if withLeft {
		l = n.children[i-1]
		r = n.children[i]
	} else {
		l = n.children[i]
		r = n.children[i+1]
	}

	keys := make([]string, 0, len(l.keys)+len(r.keys))
	keys = append(keys, l.keys...)
	if !l.leaf && withLeft {
		keys = append(keys, n.keys[i-1])
	}
	if !l.leaf && !withLeft {
		keys = append(keys, n.keys[i])
	}
	keys = append(keys, r.keys...)

	j := redistributionIndex(keys)

	// Update keys
	if l.leaf {
		l.keys = keys[:j]
		r.keys = keys[j:]
	} else {
		l.keys = keys[:j]
		r.keys = keys[j+1:]
	}
	if withLeft {
		n.keys[i-1] = keys[j]
	} else {
		n.keys[i] = keys[j]
	}

	if l.leaf {
		// Combined key locations
		keyLocations := make([]*KeyLocation, 0, len(l.keyLocations)+len(r.keyLocations))
		keyLocations = append(keyLocations, l.keyLocations...)
		keyLocations = append(keyLocations, r.keyLocations...)

		// Update key locations
		l.keyLocations = keyLocations[:j]
		r.keyLocations = keyLocations[j:]
	} else {
		// Combined children
		children := make([]*node, 0, len(l.children)+len(r.children))
		children = append(children, l.children...)
		children = append(children, r.children...)

		// Move children
		l.children = children[:j+1]
		r.children = children[j+1:]

		// Combined children
		childPageIDs := make([]uint32, 0, len(l.childPageIDs)+len(r.childPageIDs))
		childPageIDs = append(childPageIDs, l.childPageIDs...)
		childPageIDs = append(childPageIDs, r.childPageIDs...)

		// Move children pageID
		l.childPageIDs = childPageIDs[:j+1]
		r.childPageIDs = childPageIDs[j+1:]
	}
	return l, r
}

// Writes parent and left and right sibling nodes to disk
func (n *node) writeRedistribution(l *node, r *node) error {
	if err := n.writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to write parent node to disk: %w", err)
	}
	if err := l.writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to write left sibling to disk: %w", err)
	}
	if err := r.writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to write right sibling to disk: %w", err)
	}
	return nil
}

// Merges child node with their left sibling node
func (n *node) mergeWithLeft(i int) error {
	withLeft := true
	l, r := n.computeMerge(i, withLeft)
	return n.writeMerge(l, r)
}

// Merges child node with their right sibling node
func (n *node) mergeWithRight(i int) error {
	withLeft := false
	l, r := n.computeMerge(i, withLeft)
	return n.writeMerge(l, r)
}

// Mergre two sibling node into one
func (n *node) computeMerge(i int, withLeft bool) (*node, *node) { // I should test this
	var l *node
	var r *node
	var separatorKey string

	if withLeft {
		l = n.children[i-1]
		r = n.children[i]
		separatorKey = n.keys[i-1]
	} else {
		l = n.children[i]
		r = n.children[i+1]
		separatorKey = n.keys[i]
	}

	for j, keyVal := range n.keys {
		if keyVal == separatorKey {
			n.keys = append(n.keys[:j], n.keys[j+1:]...)
			break
		}
	}

	// Updates keys and key location
	if l.leaf {
		l.keys = append(l.keys, r.keys...)
		l.keyLocations = append(l.keyLocations, r.keyLocations...)
	} else {
		l.keys = append(l.keys, separatorKey)
		l.keys = append(l.keys, r.keys...)
	}

	// Updates linked list
	if l.leaf {
		l.Next = r.Next
		l.NextID = r.NextID
	}

	// Updates children
	l.children = append(l.children, r.children...)
	l.childPageIDs = append(l.childPageIDs, r.childPageIDs...)

	// Removes pointer from parent to right node
	if withLeft {
		n.children = append(n.children[:i], n.children[i+1:]...)
		n.childPageIDs = append(n.childPageIDs[:i], n.childPageIDs[i+1:]...)
	} else {
		n.children = append(n.children[:i+1], n.children[i+2:]...)
		n.childPageIDs = append(n.childPageIDs[:i+1], n.childPageIDs[i+2:]...)
	}
	return l, r
}

func (n *node) writeMerge(l *node, r *node) error {
	// Write nodes to disk
	if err := n.writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to write parent node to disk: %w", err)
	}
	if err := l.writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to write left sibling to disk: %w", err)
	}
	if err := r.deleteNodePage(); err != nil {
		return fmt.Errorf("failed to delete right node page: %w", err)
	}
	return nil
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
		return fmt.Errorf("failed to read child nodes page: %w", err)
	}
	return nil
}

// Find the key index to split page as equal as possible in half
func redistributionIndex(keys []string) int { // Should test this
	shortestDistance := math.MaxInt
	totalLen := 0
	for _, key := range keys {
		totalLen += len(key)
	}
	halfLen := totalLen / 2

	currLen := 0
	middleIndex := 0

	// Calculates index
	for i, key := range keys {
		currLen += len(key)
		distanceFromMiddle := abs(halfLen - currLen)
		if distanceFromMiddle < shortestDistance {
			middleIndex = i
			shortestDistance = distanceFromMiddle
		} else {
			break
		}
	}
	return middleIndex + 1
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
