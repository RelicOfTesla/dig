package dig2

import (
	"github.com/pkg/errors"
	"reflect"
	"sync"
)

////////////////
//
type holdTypeProvider struct {
	// implement iProvider
	mutex sync.RWMutex
	tps   map[key]bool
}

func newHoldTypeList() *holdTypeProvider {
	return &holdTypeProvider{
		tps: map[key]bool{},
	}
}

func (x *holdTypeProvider) FindValueCreator(bCtx ProviderBuildContext, t TargetKey) (TargetValueCreator, error) {
	k := toKey(t)
	if !x.hasType(k) {
		return bCtx.CallNextFindValueCreator(x, t)
	}
	return &TargetValueCreatorImpl{
		GetValue: func(gCtx GetterContext) (reflect.Value, error) {
			stores := gCtx.GetHoldValueStore()
			v, err := stores.get(t.GetType())
			if err != nil {
				return invalidReflectValue, errors.WithMessage(err, bCtx.GetPathString())
			}
			return v, nil
		},
		flatten: false,
	}, nil
}
func (x *holdTypeProvider) putType(t *key) error {
	x.mutex.Lock()
	defer x.mutex.Unlock()
	if _, has := x.tps[*t]; has {
		return errors.WithStack(ErrDuplicateProvider)
	}
	x.tps[*t] = true
	return nil
}
func (x *holdTypeProvider) hasType(t *key) bool {
	x.mutex.RLock()
	defer x.mutex.RUnlock()
	_, has := x.tps[*t]
	return has
}

/////////

type HoldValueStore struct {
	// implement GetterContext
	values map[reflect.Type]reflect.Value
}

func NewArgvHold() *HoldValueStore {
	return &HoldValueStore{
		values: nil,
	}
}

func (x *HoldValueStore) GetHoldValueStore() *HoldValueStore {
	return x
}
func (x *HoldValueStore) Clone() *HoldValueStore {
	nv := &HoldValueStore{}
	if x.values != nil {
		nv.values = map[reflect.Type]reflect.Value{}
		for k, v := range x.values {
			nv.values[k] = v
		}
	}
	return nv
}
func (x *HoldValueStore) Append(tp reflect.Type, v reflect.Value) error {
	if x.values != nil {
		if _, has := x.values[tp]; has {
			return errors.WithStack(ErrDuplicateProvider)
		}
	} else {
		x.values = make(map[reflect.Type]reflect.Value, 1)
	}
	x.values[tp] = v
	return nil
}
func (x *HoldValueStore) get(tp reflect.Type) (reflect.Value, error) {
	if x.values != nil {
		if v, has := x.values[tp]; has {
			return v, nil
		}
	}
	return invalidReflectValue, errors.WithStack(ErrNotInitHoldValue)
}
