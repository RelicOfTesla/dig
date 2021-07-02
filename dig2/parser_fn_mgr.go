package dig2

import (
	"github.com/pkg/errors"
	"reflect"
)

/////
type fnParseProviderMgr struct {
	parserList []fnParser
}
type parserContext struct {
	rootMgr *providerMgr
	added   int
	stop    bool
}
type fnParser interface {
	ParseCreator(pCtx *parserContext, creator reflect.Value, store providerStore, opt *provideOptionImpl) error
}

type providerStore interface {
	PutKvProvider(t TargetKey, p iProvider, allowReplace bool) error
}

func newFnParseProviderMgr() *fnParseProviderMgr {
	r := &fnParseProviderMgr{
		parserList: []fnParser{},
	}
	r.AddParseProvider(newNativeOutParser())
	r.AddParseProvider(newDigOutParser())
	return r
}
func (s *fnParseProviderMgr) AddParseProvider(p fnParser) {
	s.parserList = append(s.parserList, p)
}

func (s *fnParseProviderMgr) addFromCreatorFunc(creator interface{}, store providerStore, opt *provideOptionImpl, rootMgr *providerMgr) error {
	rCreator := unwrapDoubleValueOf(creator)
	if rCreator.Kind() != reflect.Func {
		return errors.WithStack(ErrMustFuncType)
	}
	tCreator := rCreator.Type()
	if tCreator.NumOut() == 0 {
		return errors.WithStack(ErrMustFuncResult)
	}
	pCtx := &parserContext{
		rootMgr: rootMgr,
		added:   0,
		stop:    false,
	}
	for i := len(s.parserList) - 1; i >= 0; i-- {
		pa := s.parserList[i]
		err := pa.ParseCreator(pCtx, rCreator, store, opt)
		if err != nil {
			return err
		}
		if pCtx.stop {
			break
		}
	}
	if pCtx.added == 0 && !pCtx.stop {
		return errors.WithStack(ErrNotFoundCreatorParser)
	}
	return nil
}
