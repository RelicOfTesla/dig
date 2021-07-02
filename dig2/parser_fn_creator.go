package dig2

import (
	"github.com/pkg/errors"
	"reflect"
)

/////////////////////////////////
type nativeOutParser struct {
	// implement fnParser
}

func newNativeOutParser() *nativeOutParser {
	return &nativeOutParser{}
}

func (x *nativeOutParser) ParseCreator(pCtx *parserContext, creator reflect.Value, store providerStore, opt *provideOptionImpl) error {
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
	for i, n := 0, outCount; i < n; i++ {
		ot := tCreator.Out(i)
		if isIgnoreType(ot) {
			continue
		}

		k := newKeyFromFull(ot, opt.Name, opt.Group)
		rp := newNativeResultPicker(i, lastIsErr)
		fnProv := newDynamicFnProvider(creator, rp, pCtx.rootMgr)
		fnProv.flatten = opt.Flatten
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
	return nil
}

func newNativeResultPicker(idx int, lastIsErr bool) resultPickerFunc {
	return func(values []reflect.Value) (reflect.Value, error) {
		if idx >= 0 && idx < len(values) && len(values) > 0 {
			if lastIsErr {
				rLast := values[len(values)-1]
				if rLast.CanInterface() {
					vLast := rLast.Interface()
					if err, ok := vLast.(error); ok && err != nil {
						return invalidReflectValue, err
					}
				}
			}
			return values[idx], nil
		}
		return invalidReflectValue, errors.WithStack(ErrInvalidOutIndex)
	}
}

var vtNativeInterfaceType = reflect.TypeOf((*interface{})(nil)).Elem()

func isIgnoreType(t reflect.Type) bool {
	if t == vtReflectValue || t == vtNativeInterfaceType {
		return true
	}
	if isErrorType(t) {
		return true
	}
	return false
}
