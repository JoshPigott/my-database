package database

import (
	"errors"
	"fmt"
	"math"
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
		if err := t.root.rootPageLink(); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
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
					return fmt.Errorf("failed to delete key: %w", err)
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
	if !nChild.isUnderflow() {
		return nil
	}

	isLeft := i != 0
	isRight := i != len(n.children)-1

	// Load children pages
	if isLeft {
		if err := n.loadChildren(i - 1); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
	}
	if isRight {
		if err := n.loadChildren(i + 1); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
	}

	if isLeft && n.needsLeftRedistribution(i) {
		if err := n.redistribution(i, true); err != nil {
			return fmt.Errorf("failed delete key: %w", err)
		}
	} else if isRight && n.needsRightRedistribution(i) {
		if err := n.redistribution(i, false); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
	} else if isLeft {
		if err := n.mergeWithLeft(i); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
	} else if isRight {
		if err := n.mergeWithRight(i); err != nil {
			return fmt.Errorf("failed to delete key: %w", err)
		}
	}
	return nil
}

// Rewrites child and right child to have the a more even split of data
func (n *node) redistribution(i int, isLeftRedistribution bool) error {
	var l *node
	var r *node
	if isLeftRedistribution {
		l = n.children[i-1]
		r = n.children[i]
	} else {
		l = n.children[i]
		r = n.children[i+1]
	}

	keys := make([]string, 0, len(l.keys)+len(r.keys))
	keys = append(keys, l.keys...)
	if !l.leaf && isLeftRedistribution {
		keys = append(keys, n.keys[i-1])
	}
	if !l.leaf && !isLeftRedistribution {
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
	if isLeftRedistribution {
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

	// Writes nodes to disk
	if err := n.writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to do redistribution: %w", err)
	}
	if err := r.writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to do redistribution: %w", err)
	}
	if err := l.writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to do redistribution: %w", err)
	}
	return nil
}

// Merges child node with their left sibling node
func (n *node) mergeWithLeft(i int) error {
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
		left.keyLocations = append(left.keyLocations, n.children[i].keyLocations...)
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
	if err := n.writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to mergre with left: %w", err)
	}
	if err := left.writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to mergre with left: %w", err)
	}
	if err := merged.deleteNodePage(); err != nil {
		return fmt.Errorf("failed to mergre with left: %w", err)
	}
	return nil
}

// Merges child node with their right sibling node
func (n *node) mergeWithRight(i int) error {
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
		n.children[i].keyLocations = append(n.children[i].keyLocations, right.keyLocations...)
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
	if err := n.writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to merge with right: %w", err)
	}
	if err := n.children[i].writeNodeToPage(); err != nil {
		return fmt.Errorf("failed to merge with right: %w", err)
	}
	if err := right.deleteNodePage(); err != nil {
		return fmt.Errorf("failed to merge with right: %w", err)
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
		return fmt.Errorf("failed to read right child: %w", err)
	}
	return nil
}

// Find the key index to split page as equal as possible in half
func redistributionIndex(keys []string) int {
	shortestDistance := math.MaxInt
	totalLen := 0
	for _, key := range keys {
		totalLen += len(key)
	}
	halfLen := totalLen / 2

	currLen := 0
	middleIndex := 0

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
