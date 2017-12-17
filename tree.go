package logtop

import (
	"github.com/timtadh/data-structures/types"
)

// copy of https://github.com/timtadh/data-structures/blob/master/tree/util.go
// but reverse order

func pop(stack []types.TreeNode) ([]types.TreeNode, types.TreeNode) {
	if len(stack) <= 0 {
		return stack, nil
	} else {
		return stack[0 : len(stack)-1], stack[len(stack)-1]
	}
}

func btn_expose_nil(node types.BinaryTreeNode) types.BinaryTreeNode {
	if types.IsNil(node) {
		return nil
	}
	return node
}

func tn_expose_nil(node types.TreeNode) types.TreeNode {
	if types.IsNil(node) {
		return nil
	}
	return node
}

func TraverseBinaryTreeInReverseOrder(node types.BinaryTreeNode) types.TreeNodeIterator {
	stack := make([]types.TreeNode, 0, 10)
	var cur types.TreeNode = btn_expose_nil(node)
	var tn_iterator types.TreeNodeIterator
	tn_iterator = func() (tn types.TreeNode, next types.TreeNodeIterator) {
		if len(stack) > 0 || cur != nil {
			for cur != nil {
				stack = append(stack, cur)
				cur = cur.(types.BinaryTreeNode).Right()
			}
			stack, cur = pop(stack)
			tn = cur
			cur = cur.(types.BinaryTreeNode).Left()
			return tn, tn_iterator
		} else {
			return nil, nil
		}
	}
	return tn_iterator
}
