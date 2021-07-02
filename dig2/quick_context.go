package dig2

import (
	"github.com/pkg/errors"
)

////////

// like context.WithValue implement quick version

type safeProviderSearchContext struct {
	// implement ProviderBuildContext
	// implement providerListGetter
	currentProvider providerIterator
	firstProvider   providerIterator
	trace           *quickCloneStack // tracePair
	path            *quickCloneStack // pathNode
	rootMgr         *providerMgr     // only value, when in findValueCreator quick get it
}

func newSafeProviderSearchContext(providers providerIterator, rootMgr *providerMgr) *safeProviderSearchContext {
	return &safeProviderSearchContext{
		currentProvider: providers,
		firstProvider:   providers,
		trace:           nil,
		path:            nil,
		rootMgr:         rootMgr,
	}
}
func (ctx *safeProviderSearchContext) clone() *safeProviderSearchContext {
	return &safeProviderSearchContext{
		currentProvider: ctx.currentProvider,
		firstProvider:   ctx.firstProvider,
		trace:           ctx.trace,
		path:            ctx.path,
		rootMgr:         ctx.rootMgr,
	}
}

func (ctx *safeProviderSearchContext) GetRootMgr() *providerMgr {
	return ctx.rootMgr
}

func (ctx *safeProviderSearchContext) GetCurrentProvider() iProvider {
	return ctx.currentProvider.Value()
}

type pathNode struct {
	k    *key
	name string
}

func (x *pathNode) String() string {
	if x.name != "" {
		return "[" + x.name + " " + x.k.String() + "]"
	} else {
		return "[" + x.k.String() + "]"
	}
}

func (ctx *safeProviderSearchContext) GetPathString() string {
	str := ""
	last := ctx.path
	for last != nil {
		v, ok := last.Value().(*pathNode)
		if ok {
			if str != "" {
				str = v.String() + "->" + str
			} else {
				str = v.String() + "."
			}
		}
		last = last.Prev()
	}
	return str
}

//

func (ctx *safeProviderSearchContext) WithApply(fnList ...WithProviderSearchContextOptFunc) (ProviderBuildContext, error) {
	info := &ctxInfo{
		raw: ctx,
		old: ctx,
		new: nil,
	}
	for _, fn := range fnList {
		err := fn(info)
		if err != nil {
			return nil, errors.Wrap(err, info.old.GetPathString())
		}
		if info.new != nil {
			info.old = info.new
		}
	}
	if info.new == nil {
		info.new = info.old
	}
	return info.new, nil
}

type WithProviderSearchContextOptFunc func(ctxInfo *ctxInfo) error

type ctxInfo struct {
	raw, old, new *safeProviderSearchContext
}

func (info *ctxInfo) getNewOrCloneNew() *safeProviderSearchContext {
	if info.new == nil {
		info.new = info.old.clone()
	}
	return info.new
}

func ApplyNotRepeatKey(k *tracePair) WithProviderSearchContextOptFunc {
	return func(ctxInfo *ctxInfo) error {
		if ctxInfo.old.trace != nil {
			if hasTreeValue(ctxInfo.old.trace, *k) {
				return errors.WithStack(ErrLoopRequire)
			}
		}
		newCtx := ctxInfo.getNewOrCloneNew()
		if newCtx.trace == nil {
			newCtx.trace = newQuickCloneStack()
		}
		newCtx.trace = newCtx.trace.PushBack(*k)
		return nil
	}
}

func hasTreeValue(x *quickCloneTreeNode, val interface{}) bool {
	last := x
	for last != nil {
		if last.Value() == val {
			return true
		}
		last = last.Prev()
	}
	return false
}

func ApplyFirstProvider() WithProviderSearchContextOptFunc {
	return func(ctxInfo *ctxInfo) error {
		newCtx := ctxInfo.getNewOrCloneNew()
		newCtx.currentProvider = newCtx.firstProvider
		return nil
	}
}
func ApplyNextProvider() WithProviderSearchContextOptFunc {
	return func(ctxInfo *ctxInfo) error {
		prev := ctxInfo.old.currentProvider.Prev()
		if prev == nil {
			return errors.WithStack(ErrNotFoundTargetProvider)
		}
		newCtx := ctxInfo.getNewOrCloneNew()
		newCtx.currentProvider = prev
		return nil
	}
}
func ApplyNewProviderList(list providerIterator) WithProviderSearchContextOptFunc {
	return func(ctxInfo *ctxInfo) error {
		newCtx := ctxInfo.getNewOrCloneNew()
		newCtx.currentProvider = list
		newCtx.firstProvider = list
		return nil
	}
}

func ApplyPathPush(tp TargetKey, name string) WithProviderSearchContextOptFunc {
	return func(ctxInfo *ctxInfo) error {
		newCtx := ctxInfo.getNewOrCloneNew()
		if newCtx.path == nil {
			newCtx.path = newQuickCloneStack()
		}
		node := &pathNode{k: toKey(tp), name: name}
		newCtx.path = newCtx.path.PushBack(node)
		return nil
	}
}

////

func (ctx *safeProviderSearchContext) CallNextFindValueCreator(historyProv iProvider, historyKey TargetKey) (TargetValueCreator, error) {
	newCtx, err := ctx.WithApply(ApplyNotRepeatKey(NewTracePair(historyProv, historyKey)), ApplyNextProvider())
	if err != nil {
		return nil, err
	}
	nextProv := newCtx.GetCurrentProvider()
	if nextProv == nil {
		return nil, errors.New("invalid provider")
	}
	return nextProv.FindValueCreator(newCtx, historyKey)
}

type tracePair struct {
	provPtr iProvider
	key     key
}

func NewTracePair(ignore iProvider, ignoreKey TargetKey) *tracePair {
	return &tracePair{
		provPtr: ignore,
		key:     *toKey(ignoreKey),
	}
}

/////////
