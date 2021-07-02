package dig2

// like context.WithValue / tree node

type quickCloneStack = quickCloneTreeNode
type quickCloneTreeNode struct {
	prev *quickCloneTreeNode
	val  interface{}
	len  uint
}

func newQuickCloneStack() *quickCloneStack {
	r := &quickCloneStack{}
	return r
}

func (x *quickCloneTreeNode) PushBack(val interface{}) *quickCloneTreeNode {
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
func (x *quickCloneTreeNode) Value() interface{} {
	return x.val
}
func (x *quickCloneTreeNode) Len() uint {
	return x.len
}
