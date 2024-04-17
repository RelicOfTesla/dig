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

	type reqArgId struct {
		id int
	}
	di.Provide(func(req *http.Request) reqArgId {
		id, err := strconv.Atoi(req.FormValue("id"))
		if err != nil {
			panic(err)
		}
		return reqArgId{
			id: id,
		}
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
	http.Handle("/test4", diHandler(func(req reqArgId) test {
		// auto response from DefaultHttpJsonApiAfterRequest
		return test{A: req.id + 4}
	}))
	srv := httptest.NewServer(nil)

	s1 := httpTestGet(t, srv.URL+"/test1?id=99")
	require.Equal(t, s1, "1,99")
	s2 := httpTestGet(t, srv.URL+"/test2")
	require.Equal(t, s2, "2")
	s3 := httpTestGet(t, srv.URL+"/test3")
	require.Equal(t, s3, `{"A":3}`)
	s4 := httpTestGet(t, srv.URL+"/test4?id=50")
	require.Equal(t, s4, `{"A":54}`)
}

func httpTestGet(t *testing.T, u string) string {
	resp, err := http.Get(u)
	require.NoError(t, err)
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(data)
}
