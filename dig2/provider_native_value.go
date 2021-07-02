package dig2

import "reflect"

type nativeValueProvider struct {
	tp  reflect.Type
	val reflect.Value
}

func NewNativeValueProvider(tp reflect.Type, val reflect.Value) *nativeValueProvider {
	return &nativeValueProvider{tp: tp, val: val}
}

func (x *nativeValueProvider) FindValueCreator(bCtx ProviderBuildContext, t TargetKey) (TargetValueCreator, error) {
	if x.tp != t.GetType() { // support interface==interface , not support interface<=>struct
		return bCtx.CallNextFindValueCreator(x, t)
	}
	return &TargetValueCreatorImpl{
		GetValue: func(context GetterContext) (value reflect.Value, e error) {
			return x.val, nil
		},
		flatten: false,
	}, nil
}
