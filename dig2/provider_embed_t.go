package dig2

import (
	"github.com/pkg/errors"
	"reflect"
)

type embedProvider struct {
	embedType  reflect.Type
	queryImpl  bool
	getValueFn any
}

func NewEmbedProvider(embedType reflect.Type, queryImpl bool, getValueFn any) *embedProvider {
	return &embedProvider{
		embedType:  embedType,
		queryImpl:  queryImpl,
		getValueFn: getValueFn,
	}
}

var vtTargetKey = reflect.TypeOf((*TargetKey)(nil)).Elem()
var vtGetterContext = reflect.TypeOf((*GetterContext)(nil)).Elem()

func (x *embedProvider) FindValueCreator(bCtx ProviderBuildContext, t TargetKey) (TargetValueCreator, error) {
	tp := t.GetType()
	if !cacheEmbedsType(tp, x.embedType) {
		yes := false
		if x.queryImpl {
			yes = cacheIsImplement(tp, x.embedType)
		}
		if !yes {
			return bCtx.CallNextFindValueCreator(x, t)
		}
	}
	if f, ok := x.getValueFn.(getValueFn); ok {
		return &TargetValueCreatorImpl{
			GetValue: f,
			flatten:  false,
		}, nil
	}
	builder := NewBuilderFromCtx(bCtx)
	err := builder.AddHoldTypeProvider(newKeyFromOnlyType(vtGetterContext))
	if err != nil {
		return nil, err
	}
	err = builder.AddKvProvider(newKeyFromOnlyType(vtTargetKey), NewNativeValueProvider(vtTargetKey, reflect.ValueOf(t)), false)
	if err != nil {
		return nil, err
	}
	firstCtx, err := bCtx.WithApply(ApplyNotRepeatKey(NewTracePair(x, t)), ApplyNewProviderList(builder.getProviders()), ApplyPathPush(t, ""))
	caller, err := builder.BuildCtx(firstCtx, x.getValueFn)
	if err != nil {
		return nil, err
	}

	return &TargetValueCreatorImpl{
		GetValue: func(gCtx GetterContext) (reflect.Value, error) {
			if err != nil {
				return invalidReflectValue, err
			}
			ctx := gCtx.GetHoldValueStore().Clone()
			err = ctx.Append(vtGetterContext, reflect.ValueOf(gCtx))
			if err != nil {
				return invalidReflectValue, err
			}
			ret, err := caller.NativeCall(ctx)
			if err != nil {
				return invalidReflectValue, err
			}
			if len(ret) == 0 {
				return invalidReflectValue, errors.New("invalid result")
			}
			v := ret[0]
			if v2, ok := v.Interface().(reflect.Value); ok {
				v = v2
			}

			return v, nil
		},
		flatten: false,
	}, nil
}
