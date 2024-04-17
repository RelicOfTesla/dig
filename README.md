* Fast dependency injection(DI) container in go (golang)
* like [uber-go/dig](https://github.com/uber-go/dig) library, but better than it :) .


# This Future
  * Split build and call, fast call speed in product.
  * Support user custom provider.
  * Support delay inject. (dynamic inject from call argument)
  * Thread safely.

## Example:
  [http_test.go](dig2/dig2_handler/http_test.go)   
  [container_test.go](dig2/container_test.go)  
  [request_di_test.go](dig2/request_di_test.go)   
  
```go

func TestHttpHandler(t *testing.T) {
	di := dig2.New()
	type test struct {
		A int
	}
	di.Provide(func() test {
		return test{A: 1}
	})

	diHandler := dig2_handler.NewHttpHandler(di, dig2_handler.HttpHandlerOpt{
		AfterRequest: dig2_handler.DefaultHttpJsonApiAfterRequest,
	})
	http.Handle("/test1", diHandler(func(resp http.ResponseWriter, a test, req *http.Request) {
		fmt.Fprintf(resp, "%v,%v", a.A, req.FormValue("id"))
	}))
	http.Handle("/test2", diHandler(func(a test) string {
		// auto response from DefaultHttpJsonApiAfterRequest
		return strconv.Itoa(a.A + 1)
	}))
	http.Handle("/test3", diHandler(func(a test) test {
		// auto response from DefaultHttpJsonApiAfterRequest
		return test{A: a.A + 2}
	}))
	srv := httptest.NewServer(nil)

	s1 := httpTestGet(t, srv.URL+"/test1?id=99")
	require.Equal(t, s1, "1,99")
	s2 := httpTestGet(t, srv.URL+"/test2")
	require.Equal(t, s2, "2")
	s3 := httpTestGet(t, srv.URL+"/test3")
	require.Equal(t, s3, `{"A":3}`)
}
```

# Some TODO
  * some uber.dig test function not cover.
  * ptr auto cast

---

