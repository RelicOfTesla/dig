package dig2

import (
	"github.com/pkg/errors"
	"math/rand"
	"reflect"
	"testing"
)

func assertErrorMatches(t *testing.T, err error, msg string, msgs ...string) {
	if err == nil {
		t.Errorf("expected error but got nil")
		return
	}
}

func RootCause(err error) error {
	for {
		e2 := errors.Unwrap(err)
		if e2 == nil {
			break
		}
		err = e2
	}
	return err
}

func IsCycleDetected(err error) bool {
	return errors.Is(err, ErrLoopRequire)
}
func setRand(r *rand.Rand) Option {
	return optionFunc(func(c *optionImpl) {
		c.rand = r
	})
}
func DryRun(dry bool) Option {
	return optionFunc(func(c *optionImpl) {
		c.dry = dry
	})
}

func DeferAcyclicVerification() Option {
	return optionFunc(func(c *optionImpl) {
		c.deferAcyclicVerification = true
	})
}
func anonymousField(t reflect.Type) reflect.StructField {
	return reflect.StructField{Name: t.Name(), Anonymous: true, Type: t}
}

func IsIn(o interface{}) bool {
	return embedsType(o, _digInType)
}
func IsOut(o interface{}) bool {
	return embedsType(o, _digOutType)
}
