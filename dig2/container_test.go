package dig2

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestDiCall_Sample(t *testing.T) {
	must := require.New(t)
	type TA struct {
		a int
	}
	called := 0
	di := New()
	must.NoError(di.Provide(func() TA {
		return TA{
			a: 100,
		}
	}))
	di.CallMust(func(ta TA) {
		must.Equal(ta.a, 100)
		called++
	})

	type TB struct {
		b int
	}
	must.NoError(di.Provide(func(ta TA) TB {
		return TB{
			b: ta.a + 2,
		}
	}))
	_, err := di.Call(func(tb TB) {
		must.Equal(tb.b, 102)
		called++
	})
	must.NoError(err)

	must.Equal(called, 2)
}

func TestDiCall_Start(t *testing.T) {
	must := require.New(t)
	di := New()
	type Ctx struct {
		a int
	}
	type TA struct {
		b int
	}
	must.NoError(di.Provide(func(ctx Ctx) TA {
		return TA{
			b: ctx.a + 1,
		}
	}))
	type TB struct {
		c int
	}
	must.NoError(di.Provide(func(ctx Ctx) TB {
		return TB{
			c: ctx.a + 10,
		}
	}))
	type TC struct {
		d int
	}
	type TD struct {
		e int
	}
	must.NoError(di.Provide(func(b TB) TC {
		return TC{
			d: b.c + 100,
		}
	}))

	called := 0

	(func() {
		var err error
		inv := di.NewInvokeBuilder()
		err = inv.AddPlaceholderFuncProvider(func(ctx Ctx) {})
		must.NoError(err)
		caller, err := inv.Build(func(c TC) TD {
			fmt.Println("call", c)
			called++
			return TD{
				e: c.d + 3,
			}
		})
		must.NoError(err)
		ret, err := caller.CastedCall(Ctx{
			a: 5,
		})
		must.NoError(err)
		must.True(len(ret) >= 1)
		fmt.Println(ret)
		d, ok := ret[0].Interface().(TD)
		must.True(ok)
		must.Equal(d.e, 118)
	})()

	must.Equal(called, 1)
}

func TestPtrNotAutoCast(t *testing.T) {
	must := require.New(t)
	c := New()
	type TA struct{}
	type TB struct{}
	err := c.Provide(func() TA {
		return TA{}
	})
	must.NoError(err)
	err = c.Provide(func() *TB {
		return &TB{}
	})
	must.NoError(err)
	called := 0
	b, err := c.Call(func(ta *TA) {
		called++
	})
	_ = b
	must.Error(err)
	b, err = c.Call(func(tb TB) {
		called++
	})
	must.Error(err)
	must.Equal(called, 0)
}

type Ia interface {
	getName() string
}
type Ca struct {
	name string
}

func (x Ca) getName() string {
	return x.name
}
func TestInterfaceCast(t *testing.T) {
	must := require.New(t)
	called := 0

	c := New()

	inv := c.NewInvokeBuilder()

	type hello struct {
		A string
	}
	err := c.Provide(func() hello {
		return hello{"hello"}
	})
	must.NoError(err)

	err = inv.AddPlaceholderFuncProvider(func(ctx Ia) {})
	must.NoError(err)

	b, err := inv.Build(func(p Ia, hello hello) {
		must.Equal(p.getName(), "aa")
		must.Equal(hello.A, "hello")
		called++
	})
	must.NoError(err)
	_, err = b.StrictCall(
		Arg{Type: reflect.TypeOf((*Ia)(nil)).Elem(), Value: &Ca{name: "aa"}},
	)
	must.NoError(err)

	type Ia2 Ia
	type Ia3 Ia
	err = c.Provide(func() Ia2 {
		return Ca{name: "a2"}
	})
	must.NoError(err)
	err = c.Provide(func() Ia3 {
		return Ca{name: "a3"}
	})
	must.NoError(err)
	_, err = c.Call(func(p2 Ia2, p3 Ia3, hello hello) {
		must.Equal(p2.getName(), "a2")
		must.Equal(p3.getName(), "a3")
		must.Equal(hello.A, "hello")
		called++
	})
	must.NoError(err)

	b, err = inv.Build(func(p0 Ia, p2 Ia2, p3 Ia3, hello hello) {
		must.FailNow("not call me")
		called++
	})
	must.NoError(err)
	_, err = b.StrictCall()
	must.Error(err)

	b, err = inv.Build(func(p0 Ia, p2 Ia2, p3 Ia3, hello hello) {
		must.Equal(p0.getName(), "bb")
		must.Equal(p2.getName(), "a2")
		must.Equal(p3.getName(), "a3")
		must.Equal(hello.A, "hello")
		called++
	})
	must.NoError(err)
	_, err = b.StrictCall(
		Arg{Type: reflect.TypeOf((*Ia)(nil)).Elem(), Value: &Ca{name: "bb"}},
	)
	must.NoError(err)

	must.Equal(called, 3)

}
