package dig2

import (
	"github.com/pkg/errors"
	"reflect"
	"sync"
)

type key struct {
	// implement TargetKey
	t     reflect.Type
	name  string
	group string
	_dbg  string
}

func (k *key) String() string {
	ret := k.t.String()
	if k.name != "" {
		ret += "(name=" + k.name + ")"
	}
	if k.group != "" {
		ret += "(group=" + k.group + ")"
	}
	return ret
}

func (k *key) GetType() reflect.Type {
	return k.t
}

var _dbg = false

func newKeyFromOnlyType(t reflect.Type) *key {
	r := &key{t: t}
	if _dbg {
		r._dbg = r.String()
	}
	return r
}
func newKeyFromFull(t reflect.Type, name string, group string) *key {
	r := &key{t: t, name: name, group: group}
	if _dbg {
		r._dbg = r.String()
	}
	return r
}
func newKeyFromTargetKey(t TargetKey) *key {
	if k, ok := t.(*key); ok {
		return k
	} else {
		return newKeyFromOnlyType(t.GetType())
	}
}
func toKey(t TargetKey) *key {
	return newKeyFromTargetKey(t)
}

////////////////

type kvProvider struct {
	// implement iProvider
	kvm   map[key]iProvider
	mutex sync.RWMutex
}

func newKvProvider() *kvProvider {
	return &kvProvider{
		kvm: map[key]iProvider{},
	}
}

///
func (s *kvProvider) FindValueCreator(bCtx ProviderBuildContext, t TargetKey) (TargetValueCreator, error) {
	prov := s.GetProvider(t)
	if prov != nil {
		return prov.FindValueCreator(bCtx, t)
	}
	return bCtx.CallNextFindValueCreator(s, t)
}
func (s *kvProvider) GetProvider(t TargetKey) iProvider {
	k := newKeyFromTargetKey(t)
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if v, has := s.kvm[*k]; has {
		return v
	}
	return nil
}
func (s *kvProvider) PutProvider(t TargetKey, p iProvider, allowReplace bool) error {
	if t == nil {
		return errors.New("invalid provider")
	}
	k := newKeyFromTargetKey(t)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !allowReplace {
		if _, ok := s.kvm[*k]; ok {
			return errors.WithStack(ErrDuplicateProvider)
		}
	}
	s.kvm[*k] = p
	return nil
}
