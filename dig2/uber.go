package dig2

import (
	"math/rand"
	"strings"
)

////////////////

var uber_special_logic = true

func init() {
	uber_special_logic = true
}

////////////////

type optionImpl struct {
	// uber test opt
	dry  bool
	rand *rand.Rand

	deferAcyclicVerification bool // provider loop check
}

func New(opts ...Option) IProviderMgr {
	opt := newOptions(opts...)
	return newProviderMgr(opt)
}

type IProviderMgr = *providerMgr

type IFindValueCreatorProvider = iProvider

func newOptions(opts ...Option) *optionImpl {
	ret := &optionImpl{}
	for _, v := range opts {
		v.applyOption(ret)
	}
	return ret
}

type optionFunc func(*optionImpl)

func (f optionFunc) applyOption(opts *optionImpl) { f(opts) }

type Option interface {
	applyOption(*optionImpl)
}

////////////////

type provideOptionImpl struct {
	Name    string
	Group   string
	Flatten bool
	//PtrCast      bool
	AllowReplace             bool
	Cache                    bool
	deferAcyclicVerification bool
}

func (x *providerMgr) Provide(f any, _opts ...ProvideOption) error {
	opt := DefaultProviderOption
	opt.deferAcyclicVerification = x.opt.deferAcyclicVerification
	if len(_opts) >= 1 {
		opt = newProvideOptions(_opts...)
	}
	return x._initFnParser.addFromCreatorFunc(f, &staticStore{x}, opt, x)
}

// ProvideOption uber.dig style
type ProvideOption interface {
	applyProvideOption(*provideOptionImpl)
}

func newProvideOptions(opts ...ProvideOption) *provideOptionImpl {
	ret := &provideOptionImpl{}
	for _, v := range opts {
		v.applyProvideOption(ret)
	}
	return ret
}

type provideOptionFunc func(*provideOptionImpl)

func (f provideOptionFunc) applyProvideOption(opts *provideOptionImpl) { f(opts) }

func Name(name string) ProvideOption {
	return provideOptionFunc(func(opts *provideOptionImpl) {
		opts.Name = name
	})
}
func Group(group string) ProvideOption {
	return provideOptionFunc(func(opts *provideOptionImpl) {
		if group != "" {
			groupList := strings.Split(group, ",")
			group = groupList[0]
			if len(groupList) >= 2 {
				opts.Flatten = groupList[1] == _tag_group_flatten
			}
		}
		opts.Group = group
	})
}

func Cache(_cache ...bool) ProvideOption {
	cache := true
	if len(_cache) > 0 {
		cache = _cache[0]
	}
	return provideOptionFunc(func(opts *provideOptionImpl) {
		opts.Cache = cache
	})
}

// Invoke uber dig.Invoke
func (x *providerMgr) Invoke(f any) error {
	ret, err := x.Call(f)
	if err != nil {
		return err
	}
	if len(ret) > 0 {
		last := ret[len(ret)-1]
		if last.CanInterface() {
			pe := last.Interface()
			if e, ok := pe.(error); ok && e != nil {
				return e
			}
		}
	}
	return err
}
