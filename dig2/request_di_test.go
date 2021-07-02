package dig2

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

type iRequestDataParser interface {
	_iRequestBase()
}

var _vtIRequest = (reflect.TypeOf((*iRequestDataParser)(nil))).Elem()

type httpReq struct {
	Url      string
	postData map[string]string
}

func fProv(req *httpReq, t TargetKey) (value reflect.Value, e error) {
	elementTp := t.GetType()
	var mustPtr bool
	if elementTp.Kind() == reflect.Ptr {
		elementTp = elementTp.Elem()
		mustPtr = true
	}
	ret := reflect.New(elementTp).Elem()

	for k, v := range req.postData {
		f := ret.FieldByName(k)
		if f.IsValid() {
			f.SetString(v)
		}
	}
	f := ret.FieldByName("NewUrl")
	if f.IsValid() {
		f.SetString(req.Url + "###")
	}

	if mustPtr {
		ret = ret.Addr()
	}
	return ret, nil
}

///

func TestRequestDi(t *testing.T) {
	must := require.New(t)

	di := New()
	di.AppendProvider(NewEmbedProvider(_vtIRequest, true, fProv))
	var err error

	builder := di.NewInvokeBuilder()
	err = builder.AddPlaceholderFuncProvider(func(*httpReq) {})
	must.NoError(err)
	called := 0

	type requestGetIndex struct {
		iRequestDataParser

		NewUrl string
		Page   string
		Hello2 string
	}
	onGetIndex := func(reqParam *requestGetIndex) {
		switch called {
		case 0:
			must.Equal(reqParam.NewUrl, "baidu.com###")
			must.Equal(reqParam.Page, "aa")
			must.Equal(reqParam.Hello2, "22")
		case 1:
			must.Equal(reqParam.NewUrl, "qq.com###")
			must.Equal(reqParam.Page, "bb")
			must.Equal(reqParam.Hello2, "33")
		default:
			must.FailNow("invalid call count")
		}
		called++
	}
	caller_onGetIndex, err := builder.Build(onGetIndex)
	must.NoError(err)

	type requestHello struct {
		iRequestDataParser

		Hello1 string
		Hello2 string
	}
	onGetHello := func(reqParam *requestHello) {
		must.Equal(reqParam.Hello1, "hello")
		must.Equal(reqParam.Hello2, "world")
		called++
	}
	caller_onGetHello, err := builder.Build(onGetHello)
	must.NoError(err)

	ret, err := caller_onGetIndex.CastedCall(&httpReq{
		Url:      "baidu.com",
		postData: map[string]string{"Page": "aa", "Hello2": "22"},
	})
	must.NoError(err)
	must.Equal(len(ret), 0)

	ret, err = caller_onGetIndex.CastedCall(&httpReq{
		Url:      "qq.com",
		postData: map[string]string{"Page": "bb", "Hello2": "33"},
	})
	must.NoError(err)
	must.Equal(len(ret), 0)

	ret, err = caller_onGetHello.CastedCall(&httpReq{
		Url:      "github.com",
		postData: map[string]string{"Hello1": "hello", "Hello2": "world"},
	})
	must.NoError(err)
	must.Equal(len(ret), 0)

	must.Equal(called, 3)
}
