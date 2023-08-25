package count

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/apache/apisix-go-plugin-runner/cmd/go-runner/plugins/testutils"
	"github.com/nft-rainbow/rainbow-settle/common/constants"
	mredis "github.com/nft-rainbow/rainbow-settle/common/redis"
	"github.com/stretchr/testify/assert"
)

func TestCountRequestFilter(t *testing.T) {

	table := []struct {
		Count            string
		ExpectStatusCode int
	}{
		{"10", 200},
		{"10010", 400},
	}

	for i, item := range table {

		pengdingCountKey := mredis.UserPendingCountKey("1", "mint")
		countKey := mredis.UserCountKey("1", "mint")
		mredis.DB().Del(context.Background(), countKey, pengdingCountKey)

		var c Count
		header := testutils.NewHttpHeader()
		header.Set(constants.RAINBOW_USER_ID_HEADER_KEY, "1")
		header.Set(constants.RAINBOW_COST_TYPE_HEADER_KEY, "mint")
		header.Set(constants.RAINBOW_COST_COUNT_HEADER_KEY, item.Count)

		r := testutils.HttpRequest{
			Header_: header,
		}

		w := httptest.NewRecorder()
		c.RequestFilter(CountConf{}, w, &r)

		body, _ := ioutil.ReadAll(w.Result().Body)
		fmt.Println(string(body))
		assert.Equal(t, item.ExpectStatusCode, w.Result().StatusCode, i)
	}
}

func TestCountRequestFilterConsequent(t *testing.T) {
	table := []struct {
		Count            string
		ExpectStatusCode int
	}{
		{"100", 200},
		{"101", 400},
	}

	pengdingCountKey := mredis.UserPendingCountKey("1", "mint")
	countKey := mredis.UserCountKey("1", "mint")
	mredis.DB().Del(context.Background(), countKey, pengdingCountKey)

	var c Count
	for i, item := range table {
		header := testutils.NewHttpHeader()
		header.Set(constants.RAINBOW_USER_ID_HEADER_KEY, "1")
		header.Set(constants.RAINBOW_COST_TYPE_HEADER_KEY, "mint")
		header.Set(constants.RAINBOW_COST_COUNT_HEADER_KEY, item.Count)

		r := testutils.HttpRequest{
			Header_: header,
		}

		w := httptest.NewRecorder()
		c.RequestFilter(CountConf{}, w, &r)

		body, _ := ioutil.ReadAll(w.Result().Body)
		fmt.Println(string(body))
		assert.Equal(t, item.ExpectStatusCode, w.Result().StatusCode, i)
	}
}
