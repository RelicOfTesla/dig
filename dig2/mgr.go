package dig2

import (
	"github.com/pkg/errors"
	"reflect"
)

////
type providerMgr struct {
	// implement providerListGetter
	opt       *optionImpl
	providers *quickCloneStack

	_initKvProvider    *kvProvider
	_initGroupProvider *sliceProvider
	//_cacheProvider     *cacheProvider

	_initFnParser *fnParseProviderMgr
}

func newProviderMgr(opt *optionImpl) *providerMgr {
	r := &providerMgr{
		opt:                opt,
		providers:          newQuickCloneStack(),
		_initKvProvider:    newKvProvider(),
		_initGroupProvider: newSliceProvider(opt.rand),
		//_cacheProvider:     newCacheProvider(),
		_initFnParser: newFnParseProviderMgr(),
	}
	r.AppendProvider(&notFoundProvider{})
	r.AppendProvider(r._initKvProvider)
	r.AppendProvider(r._initGroupProvider)
	r.AppendProvider(newDigInProvider())
	//r.AppendProvider(r._cacheProvider)

	return r
}

func (x *providerMgr) AppendProvider(prov iProvider) {
	x.providers = x.providers.PushBack(iProvider(prov))
}

func (x *providerMgr) getProviderStack() *providerStack {
	return &providerStack{x.providers}
}

///

var DefaultProviderOption = &provideOptionImpl{
	Name:                     "",
	Group:                    "",
	Flatten:                  false,
	AllowReplace:             false,
	Cache:                    false,
	deferAcyclicVerification: false,
}

////
func (x *providerMgr) CallMust(f interface{}) []reflect.Value {
	ret, err := x.Call(f)
	if err != nil {
		panic(err)
	}
	return ret
}

func (x *providerMgr) Call(f interface{}) ([]reflect.Value, error) {
	if x.opt.dry {
		return make([]reflect.Value, reflect.TypeOf(f).NumOut()), nil
	}
	builder := x.NewInvokeBuilder()
	caller, err := builder.Build(f)
	if err != nil {
		return nil, err
	}
	return caller.StrictCall()
}
func (x *providerMgr) NewInvokeBuilder() *InvokeBuilder {
	inv := newInvokeBuilder(x.getProviderStack(), x)
	return inv
}

/////

type staticStore struct {
	// implement staticStore
	*providerMgr
}

func (x *staticStore) PutKvProvider(t TargetKey, p iProvider, allowReplace bool) error {
	k := toKey(t)
	if k.group == "" {
		return x._initKvProvider.PutProvider(t, p, allowReplace)
	} else {
		return x._initGroupProvider.appendProvider(k, p)
	}
}

////////////////////////////////////////////////////

type iProvider interface {
	FindValueCreator(ctx ProviderBuildContext, t TargetKey) (TargetValueCreator, error)
}

type TargetKey interface {
	GetType() reflect.Type
	String() string
}

type ProviderBuildContext interface {
	WithApply(fnList ...WithProviderSearchContextOptFunc) (ProviderBuildContext, error)
	GetCurrentProvider() iProvider
	GetPathString() string
	CallNextFindValueCreator(historyProv iProvider, historyKey TargetKey) (TargetValueCreator, error)
	GetRootMgr() *providerMgr
}

//
type TargetValueCreator = *TargetValueCreatorImpl
type TargetValueCreatorImpl struct {
	GetValue getValueFn
	flatten  bool
}
type getValueFn func(GetterContext) (reflect.Value, error)

type GetterContext interface {
	GetHoldValueStore() *HoldValueStore
}

////////////
type notFoundProvider struct {
	// implement iProvider
}

func (*notFoundProvider) FindValueCreator(bCtx ProviderBuildContext, t TargetKey) (TargetValueCreator, error) {
	return nil, errors.Wrap(ErrNotFoundTargetProvider, bCtx.GetPathString())
}
