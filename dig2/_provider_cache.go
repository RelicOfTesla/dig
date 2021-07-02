package dig2

import (
	"reflect"
	"sync"
)

type cacheProvider struct {
	// implement iProvider
	mutex sync.RWMutex
	kv    map[key]reflect.Value
}

func newCacheProvider() *cacheProvider {
	return &cacheProvider{
		kv: map[key]reflect.Value{},
	}
}
func (x *cacheProvider) FindValueCreator(bCtx ProviderBuildContext, t TargetKey) (TargetValueCreator, error) {
	nextCtx, err := bCtx.WithApply(ApplyNotRepeatKey(NewTracePair(x, t)), ApplyNextProvider())
	if err != nil {
		return nil, err
	}
	nextProv := nextCtx.GetCurrentProvider()
	nextCreator, err := nextProv.FindValueCreator(nextCtx, t)
	if err != nil {
		return nil, err
	}
	k := toKey(t)
	return &TargetValueCreatorImpl{
		GetValue: func(gCtx GetterContext) (reflect.Value, error) {
			last, has := x.getCacheValue(k)
			if has {
				return last, nil
			}
			r, err := nextCreator.GetValue(gCtx)
			if err != nil {
				return invalidReflectValue, err
			}
			x.setCacheValue(k, r)
			return r, nil
		},
		flatten: false,
	}, nil
}
func (x *cacheProvider) getCacheValue(k *key) (reflect.Value, bool) {
	x.mutex.RLock()
	defer x.mutex.RUnlock()
	if v, has := x.kv[*k]; has {
		return v, true
	}
	return invalidReflectValue, false
}

func (x *cacheProvider) setCacheValue(k *key, v reflect.Value) {
	x.mutex.Lock()
	defer x.mutex.Unlock()
	//if _, has := x.kv[*k]; has {
	//	return errors.WithStack(ErrDuplicateProvider)
	//}
	x.kv[*k] = v
}
