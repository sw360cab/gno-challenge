package graphql

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hasura/go-graphql-client"
	"github.com/stretchr/testify/assert"
)

func TestCreateGQLStaticQuery(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mustWrite(w, `{ "data": {
			"blocks": [
				{
					"height": 0
				}
			]}
		}`)
	})
	client := graphql.NewClient("/graphql", &http.Client{Transport: localRoundTripper{handler: mux}})

	queryLastBlockBeforeTime := &LastBlockBeforeTimeQuery{}
	b := Block{
		Height: 0,
	}
	err := client.Query(context.Background(), queryLastBlockBeforeTime,
		map[string]interface{}{
			"fromHeight": b.Height,
			"toHeight":   b.Height + 1,
			"toTime":     time.Now(),
		})

	assert.NoError(t, err)
	assert.NotEmpty(t, queryLastBlockBeforeTime)
	assert.Equal(t, 1, len(queryLastBlockBeforeTime.Blocks))
}

// localRoundTripper is an http.RoundTripper that executes HTTP transactions
// by using handler directly, instead of going over an HTTP connection.
type localRoundTripper struct {
	handler http.Handler
}

func (l localRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	l.handler.ServeHTTP(w, req)
	return w.Result(), nil
}

func mustWrite(w io.Writer, s string) {
	_, err := io.WriteString(w, s)
	if err != nil {
		panic(err)
	}
}
