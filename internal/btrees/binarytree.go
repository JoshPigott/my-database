package btrees

var root *node

type node struct {
	value int
	left  *node
	right *node
}

func newNode(value int) *node {
	return &node{
		value: value,
	}
}

// Check if number is in the tree
func InTree(num int) bool {
	node := findNode(root, num)
	if num == node.value {
		return true
	} else {
		return false
	}
}

// Inserts value into the tree
func Insert(num int) {
	node := newNode(num)
	if root == nil {
		root = node
		return
	}

	parent := findNode(root, num)
	if num == parent.value {
		return
	}

	if num < parent.value {
		parent.left = node
	} else if num > parent.value {
		parent.right = node
	}
}

// Finds a node or returns its parent.
func findNode(node *node, num int) *node {
	if num == node.value {
		return node
	} else if num < node.value {
		if node.left == nil {
			return node
		} else {
			return findNode(node.left, num)
		}
	} else {
		if node.right == nil {
			return node
		} else {
			return findNode(node.right, num)
		}
	}
}
