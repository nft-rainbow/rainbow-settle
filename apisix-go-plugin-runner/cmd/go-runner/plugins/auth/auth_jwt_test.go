package auth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	thisHttp "github.com/apache/apisix-go-plugin-runner/internal/http"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	thttp "github.com/stretchr/testify/http"
)

func TestAuthFail(t *testing.T) {
	table := []struct {
		TokenLookup   string
		Authorization string
		ExpectStatus  int
		ExpectBody    string
	}{
		{
			"header: Authorization",
			"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBcHBVc2VySWQiOjEsIktZQ1R5cGUiOjIsImV4cCI6MTY4OTIxNTAxMywiaWQiOjUsIm9yaWdfaWF0IjoxNjg2NjIzMDEzfQ.SGBQXmWm6UUkElsoRXHi6CNe4GMphGsr9pqNGVAiGmg",
			401,
			"{\"code\":401,\"message\":\"Token is expired\"}",
		},
		{
			"header: Authorization",
			"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBcHBVc2VySWQiOjEsIktZQ1R5cGUiOjIsImV4cCI6MTY4OTIxNTAxMywiaWQiOjUsIm9yaWdfaWF0IjoxNjg2NjIzMDEzfQ.SGBQXmWm6UUkElsoRXHi6CNe4GMphGsr9pqNGVAiGmm",
			401,
			"{\"code\":401,\"message\":\"signature is invalid\"}",
		},
		{
			"query: Unknown",
			"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBcHBVc2VySWQiOjEsIktZQ1R5cGUiOjIsImV4cCI6MTY4OTIxNTAxMywiaWQiOjUsIm9yaWdfaWF0IjoxNjg2NjIzMDEzfQ.SGBQXmWm6UUkElsoRXHi6CNe4GMphGsr9pqNGVAiGmm",
			401,
			"{\"code\":401,\"message\":\"query token is empty\"}",
		},
	}

	for i, item := range table {
		a := &JwtAuth{}
		w := httptest.NewRecorder()

		r := thisHttp.CreateRequest([]byte{0, 0, 0, 0})
		r.Header().Set("Authorization", item.Authorization)

		a.RequestFilter(JwtAuthConf{item.TokenLookup, "rainbow-api", "local"}, w, r)
		resp := w.Result()

		body, _ := ioutil.ReadAll(resp.Body)
		assert.Equal(t, item.ExpectStatus, resp.StatusCode, i)
		assert.Equal(t, item.ExpectBody, string(body), i)
	}
}

func TestAuthOk(t *testing.T) {
	token := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBcHBVc2VySWQiOjEsIktZQ1R5cGUiOjIsImV4cCI6MTY5Mjk1NTEyMCwiaWQiOjUsIm9yaWdfaWF0IjoxNjkwMzYzMTIwfQ.B81_Ale06D1-9bEGp2R7BpOSe3oCZZaIuyaqEjrEwdA"
	a := &JwtAuth{}
	w := httptest.NewRecorder()

	r := thisHttp.CreateRequest([]byte{0, 0, 0, 0})
	r.Header().Set("Authorization", token)

	a.RequestFilter(JwtAuthConf{"header: Authorization", "rainbow-api", "local"}, w, r)
	resp := w.Result()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	assert.Equal(t, "1", r.Header().Get("x-user-id"))
}

func TestParseConf(t *testing.T) {
	value := []byte("{\"token_lookup\":\"header: Authorization\",\"app\":\"rainbow-api\",\"env\":\"local\"}")
	a := &JwtAuth{}
	conf, err := a.ParseConf(value)
	assert.NoError(t, err)
	assert.IsType(t, JwtAuthConf{}, conf)
}

func TestTmp(t *testing.T) {
	j, _ := json.Marshal(JwtAuthConf{"header: Authorization", "rainbow-api", "local"})
	fmt.Printf("%s", j)
}

func _TestAnyPath(t *testing.T) {
	g := gin.New()
	g.Any("*path", func(c *gin.Context) {})

	var w thttp.TestResponseWriter
	c, _ := gin.CreateTestContext(&w)
	url, _ := url.Parse("/dashboard")

	c.Request = &http.Request{
		URL:    url,
		Method: "GET",
	}

	g.HandleContext(c)
	fmt.Println(c.Writer.Status())
	assert.Equal(t, "/dashboard", c.FullPath())
}
