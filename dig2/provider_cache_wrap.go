package dig2

import (
	"reflect"
	"sync"
)

type cacheProviderWrap struct {
	//implement iProvider
	owner        iProvider
	once         sync.Once
	cacheResult1 reflect.Value
	cacheResult2 error
}

func newCacheProviderWrap(provider iProvider) *cacheProviderWrap {
	return &cacheProviderWrap{owner: provider}
}

func (x *cacheProviderWrap) FindValueCreator(ctx ProviderBuildContext, t TargetKey) (TargetValueCreator, error) {
	creator, err := x.owner.FindValueCreator(ctx, t)
	if err != nil {
		return nil, err
	}
	return &TargetValueCreatorImpl{
		GetValue: func(context GetterContext) (value reflect.Value, e error) {
			x.once.Do(func() {
				ret, err := creator.GetValue(context)
				x.cacheResult1, x.cacheResult2 = ret, err
			})
			return x.cacheResult1, x.cacheResult2
		},
		flatten: creator.flatten,
	}, nil
}
