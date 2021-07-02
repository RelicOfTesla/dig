package dig2

import (
	"github.com/pkg/errors"
	"reflect"
	"strconv"
)

type InvokeBuilder struct {
	ownerProviderList providerIterator
	rootMgr           *providerMgr

	scopeHoldTypeProvider *holdTypeProvider
	scopeKvProvider       *kvProvider
}

func newInvokeBuilder(owner providerIterator, mgr *providerMgr) *InvokeBuilder {
	return &InvokeBuilder{
		ownerProviderList: owner,
		rootMgr:           mgr,
	}
}

func (x *InvokeBuilder) AddHoldTypeProvider(t TargetKey) error {
	if x.scopeHoldTypeProvider == nil {
		x.scopeHoldTypeProvider = newHoldTypeList()
	}
	return x.scopeHoldTypeProvider.putType(toKey(t))
}
func (x *InvokeBuilder) AddPlaceholderFuncProvider(f interface{}) error {
	rf := unwrapDoubleValueOf(f)
	tf := rf.Type()
	inCount := tf.NumIn()
	if tf.IsVariadic() && inCount > 0 {
		inCount--
	}
	for i := 0; i < inCount; i++ {
		f := tf.In(i)
		k := newKeyFromOnlyType(f)
		err := x.AddHoldTypeProvider(k)
		if err != nil {
			return err
		}
	}
	return nil
}
func (x *InvokeBuilder) AddKvProvider(t TargetKey, provider iProvider, allowReplace bool) error {
	if x.scopeKvProvider == nil {
		x.scopeKvProvider = newKvProvider()
	}
	return x.scopeKvProvider.PutProvider(t, provider, allowReplace)
}

func (x *InvokeBuilder) newSearchCtx() ProviderBuildContext {
	return newSafeProviderSearchContext(x.getProviders(), x.rootMgr)
}
func (x *InvokeBuilder) getProviders() providerIterator {
	if x.scopeHoldTypeProvider == nil {
		x.scopeHoldTypeProvider = newHoldTypeList()
	}
	if x.scopeKvProvider == nil {
		x.scopeKvProvider = newKvProvider()
	}
	stack := x.ownerProviderList.(*providerStack).quickCloneStack
	stack = stack.PushBack(iProvider(x.scopeHoldTypeProvider))
	stack = stack.PushBack(iProvider(x.scopeKvProvider))
	return &providerStack{stack}
}

func (x *InvokeBuilder) BuildCtx(ctx ProviderBuildContext, f interface{}) (*InvokerCaller, error) {
	rf := unwrapDoubleValueOf(f)
	if rf.Kind() != reflect.Func {
		return nil, errors.Wrap(ErrMustFuncType, ctx.GetPathString())
	}
	tf := rf.Type()
	inCount := tf.NumIn()
	if tf.IsVariadic() && inCount > 0 {
		inCount--
	}
	caller := &InvokerCaller{
		rvFn:     rf,
		in:       make([]reflect.Value, inCount),
		creator:  make([]TargetValueCreator, inCount),
		hasValue: make([]bool, inCount),
	}
	for i := 0; i < inCount; i++ {
		t := tf.In(i)
		inK := newKeyFromOnlyType(t)
		firstCtx, err := ctx.WithApply(ApplyFirstProvider(), ApplyPathPush(inK, "_"+strconv.Itoa(i+1)))
		if err != nil {
			return nil, err
		}
		firstProv := firstCtx.GetCurrentProvider()
		creator, err := firstProv.FindValueCreator(firstCtx, inK)
		if err != nil {
			return nil, err
		}
		caller.creator[i] = creator
	}
	return caller, nil
}

func (x *InvokeBuilder) Build(f interface{}) (*InvokerCaller, error) {
	return x.BuildCtx(x.newSearchCtx(), f)
}

/////
type InvokerCaller struct {
	rvFn     reflect.Value
	in       []reflect.Value
	creator  []TargetValueCreator
	hasValue []bool // todo:cache?or
}

type CallContext = GetterContext

func (x *InvokerCaller) NativeCall(cCtx CallContext) ([]reflect.Value, error) {
	for i, n := 0, len(x.in); i < n; i++ {
		if !x.hasValue[i] {
			creator := x.creator[i]
			in, err := creator.GetValue(cCtx)
			if err != nil {
				return nil, err
			}
			x.in[i] = in
			//x.hasValue[i] = true //cache // DONT CACHE.
		}
	}
	ret := x.rvFn.Call(x.in)
	return ret, nil
}

func NewArgvHoldFromStrictArg(argv ...Arg) (*HoldValueStore, error) {
	valueHolder := NewArgvHold()
	for _, v := range argv {
		err := valueHolder.Append(v.Type, reflect.ValueOf(v.Value))
		if err != nil {
			return nil, err
		}
	}
	return valueHolder, nil
}
func NewArgvHoldFromCastArg(argv ...interface{}) (*HoldValueStore, error) {
	valueHolder := NewArgvHold()
	for _, v := range argv {
		rv := unwrapDoubleValueOf(v)
		err := valueHolder.Append(rv.Type(), rv)
		if err != nil {
			return nil, err
		}
	}
	return valueHolder, nil
}

type Arg struct {
	Type  reflect.Type
	Value interface{}
}

// 注意argv...传入interface会自动转成struct，会导致与interface类型的声明不匹配，建议外部使用NewArgvHold().Append()
// fix interface auto cast to struct bug. now, use raw type
func (x *InvokerCaller) StrictCall(argv ...Arg) ([]reflect.Value, error) {
	valueHolder, err := NewArgvHoldFromStrictArg(argv...)
	if err != nil {
		return nil, err
	}
	return x.NativeCall(valueHolder)
}

// not auto cast XXX struct to XX interface
func (x *InvokerCaller) CastedCall(argv ...interface{}) ([]reflect.Value, error) {
	valueHolder, err := NewArgvHoldFromCastArg(argv...)
	if err != nil {
		return nil, err
	}
	return x.NativeCall(valueHolder)
}

/////

func NewBuilderFromCtx(ctx ProviderBuildContext) *InvokeBuilder {
	if c, ok := ctx.(*safeProviderSearchContext); ok {
		return newInvokeBuilder(c.firstProvider, ctx.GetRootMgr())
	}
	root := ctx.GetRootMgr()
	inv := newInvokeBuilder(root.getProviderStack(), root)
	return inv
}
func NewInvokeFromCtx(ctx ProviderBuildContext, nowProvider iProvider, t TargetKey, fn interface{}) (*InvokerCaller, *InvokeBuilder, error) {
	builder := NewBuilderFromCtx(ctx)
	nextCtx, err := ctx.WithApply(ApplyFirstProvider(), ApplyNotRepeatKey(NewTracePair(nowProvider, t)), ApplyPathPush(t, ""))
	if err != nil {
		return nil, nil, err
	}
	caller, err := builder.BuildCtx(nextCtx, fn)
	return caller, builder, err
}
func NewCallerFromCtx(ctx ProviderBuildContext, nowProvider iProvider, t TargetKey, fn interface{}) (*InvokerCaller, error) {
	caller, _, err := NewInvokeFromCtx(ctx, nowProvider, t, fn)
	return caller, err
}
