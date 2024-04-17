package dig2

// like context.WithValue / tree node

type quickCloneStack = quickCloneTreeNode
type quickCloneTreeNode struct {
	prev *quickCloneTreeNode
	val  any
	len  uint
}

func newQuickCloneStack() *quickCloneStack {
	r := &quickCloneStack{}
	return r
}

func (x *quickCloneTreeNode) PushBack(val any) *quickCloneTreeNode {
	newNode := &quickCloneTreeNode{
		prev: x,
		val:  val,
		len:  x.len + 1,
	}
	return newNode
}
func (x *quickCloneTreeNode) Prev() *quickCloneTreeNode {
	return x.prev
}
func (x *quickCloneTreeNode) Value() any {
	return x.val
}
func (x *quickCloneTreeNode) Len() uint {
	return x.len
}
