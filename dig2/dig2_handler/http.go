package dig2_handler

import (
	"encoding/json"
	"github.com/RelicOfTesla/dig/dig2"
	"net/http"
	"reflect"
)

func NewHttpHandler(di dig2.IProviderMgr, _opts ...HttpHandlerOpt) func(f any) http.HandlerFunc {
	opt := HttpHandlerOpt{}
	if len(_opts) > 0 {
		opt = _opts[0]
	}

	builder := di.NewInvokeBuilder()
	if err := builder.AddPlaceholderFuncProvider(http.HandlerFunc(nil)); err != nil {
		panic(err)
	}
	if opt.InitBind != nil {
		opt.InitBind(builder)
	}
	return func(f any) http.HandlerFunc {
		caller, err := builder.Build(f)
		if err != nil {
			panic(err)
		}
		return func(resp http.ResponseWriter, req *http.Request) {
			if opt.BeforeRequest != nil {
				if opt.BeforeRequest(resp, req) {
					return
				}
			}
			defer func() {
				if e := recover(); e != nil {
					if err, ok := e.(error); ok && err != nil {
						if opt.AfterRequest != nil {
							opt.AfterRequest(resp, req, nil, err)
						} else {
							panic(e)
						}
					} else {
						panic(e)
					}
				}
			}()
			ret, err := caller.StrictCall(
				//dig2.ArgFrom[http.ResponseWriter](resp),
				//dig2.ArgFrom[*http.Request](req),
				dig2.Arg{Type: vtHttpResponseWriter, Value: resp},
				dig2.Arg{Type: vtHttpRequest, Value: req},
			)
			if opt.AfterRequest != nil {
				opt.AfterRequest(resp, req, ret, err)
			}
		}
	}
}

var vtHttpResponseWriter = dig2.TypeFor[http.ResponseWriter]()
var vtHttpRequest = dig2.TypeFor[*http.Request]()

type HttpHandlerOpt struct {
	InitBind      func(builder *dig2.InvokeBuilder)
	BeforeRequest func(resp http.ResponseWriter, req *http.Request) bool
	AfterRequest  func(resp http.ResponseWriter, req *http.Request, results []reflect.Value, err error)
}

var DefaultHttpJsonApiOnErr = func(resp http.ResponseWriter, req *http.Request, results []reflect.Value, err error) {
	if err != nil {
		resp.Write([]byte(err.Error()))
		//fmt.Fprintf(resp, "%e", err)
	}
}

func DefaultHttpJsonApiAfterRequest(resp http.ResponseWriter, req *http.Request, results []reflect.Value, err error) {
	if len(results) > 0 && err == nil {
		last := results[len(results)-1]
		if last.CanInterface() {
			if retErr, ok := last.Interface().(error); ok && retErr != nil {
				err = retErr
			}
		}
	}
	if err != nil {
		DefaultHttpJsonApiOnErr(resp, req, results, err)
		return
	}

	if len(results) > 0 {
		r0 := results[0]
		switch r0.Kind() {
		case reflect.String:
			resp.Write([]byte(r0.String()))
		default:
			if r0.CanInterface() {
				p := r0.Interface()
				switch v := p.(type) {
				case interface {
					Dispatch(resp http.ResponseWriter)
				}:
					v.Dispatch(resp)
				case interface {
					String() string
				}:
					resp.Write([]byte(v.String()))
				default:
					cb, err := json.Marshal(p)
					if err != nil {
						DefaultHttpJsonApiOnErr(resp, req, results, err)
					} else {
						resp.Write(cb)
					}
				}
			}
		}
	}
}
