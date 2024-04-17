package dig2_handler_test

import (
	"fmt"
	"github.com/RelicOfTesla/dig/dig2"
	"github.com/RelicOfTesla/dig/dig2/dig2_handler"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestHttpHandler(t *testing.T) {
	di := dig2.New()
	type test struct {
		A int
	}
	di.Provide(func() test {
		return test{A: 1}
	})

	diHandler := dig2_handler.NewHttpHandler(di, dig2_handler.HttpHandlerOpt{
		AfterResult: dig2_handler.DefaultHttpJsonApiAfterResult,
	})
	http.Handle("/test1", diHandler(func(resp http.ResponseWriter, a test) {
		fmt.Fprintf(resp, "%v", a.A)
	}))
	http.Handle("/test2", diHandler(func(a test) string {
		// auto response from DefaultHttpJsonApiAfterResult
		return strconv.Itoa(a.A + 1)
	}))
	http.Handle("/test3", diHandler(func(a test) test {
		// auto response from DefaultHttpJsonApiAfterResult
		return test{A: a.A + 2}
	}))
	srv := httptest.NewServer(nil)

	s1 := httpTestGet(t, srv.URL+"/test1")
	require.Equal(t, s1, "1")
	s2 := httpTestGet(t, srv.URL+"/test2")
	require.Equal(t, s2, "2")
	s3 := httpTestGet(t, srv.URL+"/test3")
	require.Equal(t, s3, `{"A":3}`)
}

func httpTestGet(t *testing.T, u string) string {
	resp, err := http.Get(u)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(data)
}
