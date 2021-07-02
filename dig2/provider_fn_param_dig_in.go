package dig2

import (
	"github.com/pkg/errors"
	"reflect"
	"strconv"
)

//
type digSentinelIn interface {
	_digSentinelIn()
}
type In struct{ digSentinelIn }

var _digInType = reflect.TypeOf(In{})

type digInProvider struct {
	// implement iProvider
}

func newDigInProvider() *digInProvider {
	return &digInProvider{}
}

const _tag_optional = "optional"

func (x *digInProvider) FindValueCreator(bCtx ProviderBuildContext, t TargetKey) (TargetValueCreator, error) {
	tp := t.GetType()
	if !cacheEmbedsType(tp, _digInType) {
		return bCtx.CallNextFindValueCreator(x, t)
	}
	if uber_special_logic {
		if cacheEmbedsType(unwrapTypePtr(tp), _digOutType) || cacheEmbedsType(unwrapTypePtr(tp), _digOutPtrType) {
			return nil, errors.Wrap(ErrUBerDefine, bCtx.GetPathString())
		}
	}
	builderList := make([]TargetValueCreator, tp.NumField())
	for i, n := 0, tp.NumField(); i < n; i++ {
		ft := tp.Field(i)
		if ft.Type == _digInType {
			continue
		}
		k := newKeyFromStructField(ft)
		firstCtx, err := bCtx.WithApply(ApplyNotRepeatKey(NewTracePair(x, t)), ApplyFirstProvider(), ApplyPathPush(k, ft.Name))
		if err != nil {
			return nil, err
		}
		if uber_special_logic && ft.PkgPath != "" { // UnExport field (lower case member)
			return nil, errors.Wrap(ErrUnexportedField, firstCtx.GetPathString())
		}

		firstProv := firstCtx.GetCurrentProvider()
		creator, err := firstProv.FindValueCreator(firstCtx, k)
		if err != nil {
			optYes := false
			if uber_special_logic && k.group != "" {
				// 有group时，允许In 里的group没有provide声明，一个很奇怪的设计
				optYes = true
			}
			if !optYes {
				optional := ft.Tag.Get(_tag_optional)
				if optional != "" {
					opt, _ := strconv.ParseBool(optional)
					optYes = opt
				}
			}
			if !optYes {
				return nil, err
			}
			creator = nil
		}
		builderList[i] = creator
	}
	return &TargetValueCreatorImpl{
		GetValue: func(gCtx GetterContext) (reflect.Value, error) {
			ret := reflect.New(tp).Elem()
			for i, creator := range builderList {
				if creator != nil {
					r, err := creator.GetValue(gCtx)
					if err != nil {
						return invalidReflectValue, err
					}
					ff := ret.Field(i)
					if !ff.CanSet() {
						return invalidReflectValue, errors.WithStack(ErrCanNotSetFieldValue)
					}
					ff.Set(r)
				}
			}
			return ret, nil
		},
		flatten: false,
	}, nil
}
