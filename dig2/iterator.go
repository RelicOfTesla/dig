package dig2

type providerIterator interface {
	Prev() providerIterator
	Value() iProvider
}
type providerStack struct {
	*quickCloneStack
}

func (x *providerStack) Prev() providerIterator {
	prev := x.quickCloneStack.Prev()
	if prev != nil {
		return &providerStack{prev}
	}
	return nil
}

func (x *providerStack) Value() iProvider {
	p := x.quickCloneStack.Value()
	if prov, ok := p.(iProvider); ok && prov != nil {
		return prov
	}
	return nil
}
