package plugins

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeverForCheckId(t *testing.T) {
	err := http.ListenAndServe(":8086", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// rID := r.Header.Get("ID")
		xRID := r.Header.Get("X-Request-Id")
		// w.Header().Set("rID", rID)
		w.Header().Set("xRID", xRID)
		// w.Write([]byte(fmt.Sprintf("rID=%s", rID)))
		w.Write([]byte(fmt.Sprintf("xRID=%s", xRID)))
	}))
	assert.NoError(t, err)
}

func BenchmarkCheckId(b *testing.B) {
	var w sync.WaitGroup
	for i := 0; i < 50; i++ {
		w.Add(1)
		go func() {
			resp, err := http.Get("http://127.0.0.1:9080")
			if err != nil {
				return
			}
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Println(string(body))
			w.Done()
		}()
	}
	w.Wait()
}
