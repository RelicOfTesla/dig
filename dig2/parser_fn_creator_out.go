package dig2

import (
	"github.com/pkg/errors"
	"reflect"
	"strings"
)

/////////

//
type digSentinelOut interface {
	digSentinelOut()
}
type Out struct{ digSentinelOut }

var _digOutType = reflect.TypeOf(Out{})
var _digOutPtrType = reflect.TypeOf(&Out{})

////

/////////////////////////////////
type digOutParser struct {
}

func newDigOutParser() *digOutParser {
	r := &digOutParser{}
	return r
}
func (x *digOutParser) ParseCreator(pCtx *parserContext, creator reflect.Value, store providerStore, opt *provideOptionImpl) error {
	tCreator := creator.Type()
	if tCreator.Kind() != reflect.Func {
		return errors.WithStack(ErrMustFuncType)
	}
	outCount := tCreator.NumOut()
	if outCount <= 0 {
		return errors.WithStack(ErrMustFuncResult)
	}
	lastIsErr := isErrorType(tCreator.Out(outCount - 1))
	if lastIsErr {
		// ignore last error
		outCount--
	}
	for i := 0; i < outCount; i++ {
		ot := tCreator.Out(i)
		if uber_special_logic {
			// 不允许 Out里有In，和使用*Out格式，很奇怪的设计
			if cacheEmbedsType(unwrapTypePtr(ot), _digInType) {
				return errors.Wrap(ErrUBerDefine, "in Out struct has field In")
			}
			if cacheEmbedsType(unwrapTypePtr(ot), _digOutPtrType) {
				return errors.Wrap(ErrUBerDefine, "must use native Out, don't use *out")
			}
		}
		if cacheEmbedsType(unwrapTypePtr(ot), _digOutType) {
			if uber_special_logic {
				if ot.Kind() == reflect.Ptr {
					// 不允许 包含Out的provide 使用*struct 返回，很奇怪的设计
					return errors.Wrap(ErrUBerDefine, "must use native Out, don't use *out")
				}
				if opt.Name != "" {
					return errors.Wrap(ErrUBerDefine, "cannot specify a name for Out objects.")
				}
			}
			err := x.parseDigOutElement(pCtx, ot, i, lastIsErr, creator, store, opt)
			if err != nil {
				return err
			}
			pCtx.stop = true
		}
	}
	const delayCheck = false
	if uber_special_logic && !delayCheck {
		for i := 0; i < tCreator.NumIn(); i++ {
			in := tCreator.In(i)
			if cacheEmbedsType(unwrapTypePtr(in), _digOutType) || cacheEmbedsType(unwrapTypePtr(in), _digOutPtrType) {
				return errors.Wrap(ErrUBerDefine, "in In struct has field Out")
			}
		}
	}
	return nil
}
func (x *digOutParser) parseDigOutElement(pCtx *parserContext, outType reflect.Type, outIndex int, lastIsErr bool,
	creator reflect.Value, store providerStore, opt *provideOptionImpl) (e error) {

	type element struct {
		tp         reflect.Type
		fieldIndex [][]int
	}
	idle := make([]element, 0, 1)
	type fieldIndexs = []int
	idle = append(idle, element{
		tp:         outType,
		fieldIndex: []fieldIndexs{},
	})
	for len(idle) > 0 {
		now := idle[0]
		idle = idle[1:]
		if now.tp.Kind() != reflect.Struct {
			// ignore interface
			continue
		}
		for i, n := 0, now.tp.NumField(); i < n; i++ {
			ft := now.tp.Field(i)
			if ft.Type == _digOutType {
				continue
			}
			curIdx := make([]fieldIndexs, 0, len(now.fieldIndex)+1)
			curIdx = append(curIdx, now.fieldIndex...)
			curIdx = append(curIdx, ft.Index)
			k, flatten := newKeyFromStructFieldEx(ft)
			if ft.Type.Kind() == reflect.Struct && cacheEmbedsType(ft.Type, _digOutType) {
				if k.name != "" {
					return errors.Wrap(ErrUBerDefine, "cannot specify a name for Out objects.")
				}
				e := element{
					tp:         ft.Type,
					fieldIndex: curIdx,
				}
				idle = append(idle, e)
				continue
			}
			//if flatten {
			//	if ft.Type.Kind() != reflect.Slice && ft.Type.Kind() != reflect.Array && ft.Type != vtReflectValue && ft.Type != vtNativeInterfaceType {
			//		return errors.Wrap(ErrUBerDefine, "group+flatten must slice/array define")
			//	}
			//}

			rp := newDigOutResultPicker(outIndex, lastIsErr, curIdx)
			fnProv := newDynamicFnProvider(creator, rp, pCtx.rootMgr)
			fnProv.flatten = flatten
			var iProv iProvider
			iProv = fnProv
			if opt.Cache {
				iProv = newCacheProviderWrap(fnProv)
			}
			err := store.PutKvProvider(k, iProv, opt.AllowReplace)
			if err != nil {
				return err
			}
			if !opt.deferAcyclicVerification {
				_, err = fnProv.checkBuild(newSafeProviderSearchContext(pCtx.rootMgr.getProviderStack(), pCtx.rootMgr), k)
				if err != nil && !errors.Is(err, ErrNotFoundTargetProvider) {
					// todo: not revert store value!
					return err
				}
			}
			pCtx.added++
		}
	}

	return nil
}

const _tag_name = "name"
const _tag_group = "group"
const _tag_group_flatten = "flatten"

func newKeyFromStructField(ft reflect.StructField) *key {
	k, _ := newKeyFromStructFieldEx(ft)
	return k
}

func newKeyFromStructFieldEx(ft reflect.StructField) (*key, bool) {
	var flatten bool
	name := ft.Tag.Get(_tag_name)
	group := ft.Tag.Get(_tag_group)
	if group != "" {
		groupList := strings.Split(group, ",")
		group = groupList[0]
		if len(groupList) >= 2 {
			flatten = groupList[1] == _tag_group_flatten
		}
	}
	k := newKeyFromFull(ft.Type, name, group)
	return k, flatten
}

func newDigOutResultPicker(outIdx int, lastIsErr bool, fieldIdxList [][]int) resultPickerFunc {
	pickOut := newNativeResultPicker(outIdx, lastIsErr)
	return func(values []reflect.Value) (reflect.Value, error) {
		ret, err := pickOut(values)
		if err != nil {
			return ret, err
		}
		for _, idx := range fieldIdxList {
			ret = ret.FieldByIndex(idx)
		}
		return ret, nil
	}
}
