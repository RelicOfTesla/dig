package dig2

import (
	"github.com/pkg/errors"
	"math/rand"
	"reflect"
	"sync"
	"time"
)

type sliceProvider struct {
	// implement iProvider
	kv          map[key][]iProvider
	mutex       sync.RWMutex
	rand        *rand.Rand
	disableRand bool
}

func newSliceProvider(rand *rand.Rand) *sliceProvider {
	return &sliceProvider{
		rand: rand,
		kv:   map[key][]iProvider{},
	}
}
func (x *sliceProvider) getRand() *rand.Rand {
	if x.rand == nil {
		x.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return x.rand
}
func (x *sliceProvider) FindValueCreator(bCtx ProviderBuildContext, sliceKeyType TargetKey) (TargetValueCreator, error) {
	sliceKey := toKey(sliceKeyType)
	if sliceKey.group == "" {
		return bCtx.CallNextFindValueCreator(x, sliceKey)
	}
	if sliceKey.t.Kind() != reflect.Slice && sliceKey.t.Kind() != reflect.Array {
		return nil, errors.Wrap(ErrUBerDefine, bCtx.GetPathString())
	}

	elementKey := newKeyFromFull(sliceKey.t.Elem(), sliceKey.name, sliceKey.group)
	providerList := x.getProviderList(elementKey)
	sliceProviderList := x.getProviderList(sliceKey)
	if len(providerList)+len(sliceProviderList) == 0 {
		return bCtx.CallNextFindValueCreator(x, sliceKey)
	}

	creatorList := make([]TargetValueCreator, 0, len(providerList)+len(sliceProviderList))
	for _, prov := range providerList {
		creator, err := prov.FindValueCreator(bCtx, elementKey)
		if err != nil {
			return nil, err
		}
		creatorList = append(creatorList, creator)
	}
	for _, prov := range sliceProviderList {
		creator, err := prov.FindValueCreator(bCtx, sliceKey)
		if err != nil {
			return nil, err
		}
		if !creator.flatten {
			return nil, errors.Wrap(ErrUBerDefine, "group+flatten must slice/array define")
		}
		creatorList = append(creatorList, creator)
	}

	return &TargetValueCreatorImpl{
		GetValue: func(gCtx GetterContext) (reflect.Value, error) {
			items := make([]reflect.Value, 0, len(creatorList))
			for _, creator := range creatorList {
				v, err := creator.GetValue(gCtx)
				if err != nil {
					return invalidReflectValue, err
				}
				if creator.flatten && (v.Kind() == reflect.Array || v.Kind() == reflect.Slice) {
					for x, n := 0, v.Len(); x < n; x++ {
						items = append(items, v.Index(x))
					}
				} else {
					items = append(items, v)
				}
			}
			if !x.disableRand {
				items = shuffledCopy(x.getRand(), items)
			}

			rv := reflect.MakeSlice(sliceKey.t, len(items), len(items))
			for i, v := range items {
				rv.Index(i).Set(v)
			}
			return rv, nil
		},
		flatten: false,
	}, nil
}
func (x *sliceProvider) getProviderList(elementKey *key) []iProvider {
	x.mutex.RLock()
	defer x.mutex.RUnlock()
	list, has := x.kv[*elementKey]
	if !has {
		return nil
	}
	return list
}
func (x *sliceProvider) appendProvider(elementKey *key, provider iProvider) error {
	if elementKey.group == "" {
		return errors.New("not init group key")
	}
	x.mutex.Lock()
	defer x.mutex.Unlock()
	x.kv[*elementKey] = append(x.kv[*elementKey], provider)
	return nil
}

func shuffledCopy(rand *rand.Rand, items []reflect.Value) []reflect.Value {
	newItems := make([]reflect.Value, len(items))
	for i, j := range rand.Perm(len(items)) {
		newItems[i] = items[j]
	}
	return newItems
}
