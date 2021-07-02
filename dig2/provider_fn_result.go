package dig2

import (
	"reflect"
)

type dynamicFnProvider struct {
	// implement iProvider
	creator      reflect.Value
	resultPicker resultPickerFunc
	rootMgr      *providerMgr
	flatten      bool

	caller *InvokerCaller
}
type resultPickerFunc func([]reflect.Value) (reflect.Value, error)

func newDynamicFnProvider(creator reflect.Value, resultPicker resultPickerFunc, rootMgr *providerMgr) *dynamicFnProvider {
	return &dynamicFnProvider{
		creator:      creator,
		resultPicker: resultPicker,
		rootMgr:      rootMgr,
	}
}

func (x *dynamicFnProvider) checkBuild(ctx ProviderBuildContext, k TargetKey) (TargetValueCreator, error) {
	if x.caller == nil {
		nextCtx, err := ctx.WithApply(ApplyNotRepeatKey(NewTracePair(x, k)), ApplyNextProvider())
		if err != nil {
			return nil, err
		}

		inv := NewBuilderFromCtx(ctx)
		caller, err := inv.BuildCtx(nextCtx, x.creator)
		if err != nil {
			return nil, err
		}
		x.caller = caller
	}
	return nil, nil
}

func (x *dynamicFnProvider) FindValueCreator(bCtx ProviderBuildContext, k TargetKey) (TargetValueCreator, error) {
	_, err := x.checkBuild(bCtx, k)
	if err != nil {
		return nil, err
	}
	return &TargetValueCreatorImpl{
		GetValue: func(gCtx GetterContext) (reflect.Value, error) {
			ret, err := x.caller.NativeCall(gCtx)
			if err != nil {
				return invalidReflectValue, err
			}
			return x.resultPicker(ret)
		},
		flatten: x.flatten,
	}, nil
}
