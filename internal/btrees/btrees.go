package btrees

import (
	"fmt"
	"slices"
	"sort"
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
	keys     []int
	children []*node
	Next     *node
	leaf     bool
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

// A wrapper for insert to handle spliting of the root
func (t *BTree) Insert(key int) {
	splitResult, left, right := insert(t.root, key)
	if splitResult != nil {
		newRoot := newNode(false)
		newRoot.addKey(*splitResult, left, right)
		t.root = newRoot
	}
}

// Returns if key in tree and key
func (t *BTree) FindKey(key int) (*node, bool) {
	n := t.root
	if slices.Contains(n.keys, key) {
		return n, true
	}
	i := 0
	for n.leaf == false && i < 5 {
		n, _ = n.findChild(key)
		if slices.Contains(n.keys, key) {
			return n, true
		}
		i += 1
	}
	return nil, false
}

/*
Recursively calls itself to create a call stack,
adds key at leaf, and works back up splitting if needed
*/
func insert(n *node, key int) (*int, *node, *node) {
	if n.leaf {
		splitResult, left, right := n.addKey(key, nil, nil)
		return splitResult, left, right
	}

	child, _ := n.findChild(key)
	splitResult, left, right := insert(child, key)

	if splitResult != nil {
		splitResult, left, right = n.addKey(*splitResult, left, right)
	}
	return splitResult, left, right
}

func (n *node) addKey(num int, left *node, right *node) (*int, *node, *node) {
	var splitResult *int
	// Checks if key in node already
	if n.leaf == true && slices.Contains(n.keys, num) {
		return nil, nil, nil
	}

	i := sort.Search(len(n.keys), func(i int) bool {
		return num < n.keys[i]
	})

	n.keys = append(n.keys, 0)
	copy(n.keys[i+1:], n.keys[i:])
	n.keys[i] = num

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
		return leftChildKey < n.keys[i]
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
func (n *node) split() (*int, *node, *node) {
	right := newNode(true)

	middlekeyIndex := maxKeys / 2
	middlekey := &n.keys[middlekeyIndex]

	// Updates keys
	// If leaf spilt right key stays in leaf
	if n.leaf == false {
		rightStart := middlekeyIndex + 1
		right.keys = make([]int, len(n.keys[rightStart:]))
		copy(right.keys, n.keys[rightStart:])
	} else {
		rightStart := middlekeyIndex
		right.keys = make([]int, len(n.keys[rightStart:]))
		copy(right.keys, n.keys[rightStart:])
	}
	n.keys = n.keys[:middlekeyIndex]

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
func (n *node) findChild(key int) (*node, int) {
	for i := range len(n.keys) {
		if key < n.keys[i] {
			return n.children[i], i
		}
	}
	return n.children[len(n.keys)], len(n.keys)
}

// Prints out the btree structure for debugging purposes
func (t *BTree) CheckStructure(num int) {
	fmt.Println("Number:", num)
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
