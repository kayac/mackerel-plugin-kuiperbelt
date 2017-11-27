package kuiperbelt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphDefinition(t *testing.T) {
	var kuiperbelt Plugin

	graphdef := kuiperbelt.GraphDefinition()
	if len(graphdef) != 2 {
		t.Errorf("GetTempfilename: %d should be 2", len(graphdef))
	}
}

var statsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(struct {
		Connections        int64 `json:"connections"`
		TotalConnections   int64 `json:"total_connections"`
		TotalMessages      int64 `json:"total_messages"`
		ConnectErrors      int64 `json:"connect_errors"`
		MessageErrors      int64 `json:"message_errors"`
		ClosingConnections int64 `json:"closing_connections"`
	}{
		Connections:        10,
		TotalConnections:   123,
		TotalMessages:      9801,
		ConnectErrors:      3,
		MessageErrors:      42,
		ClosingConnections: 2,
	})
})

func TestParse(t *testing.T) {
	var kuiperbelt Plugin
	mux := http.NewServeMux()
	mux.HandleFunc("/stats", statsHandler)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	kuiperbelt.Target = u.Host

	stat, err := kuiperbelt.FetchMetrics()
	fmt.Println(stat)
	assert.Nil(t, err)
	// Kuiperbelt Stats
	assert.EqualValues(t, stat["conn.current"], 10)
	assert.EqualValues(t, stat["conn.total"], 123)
	assert.EqualValues(t, stat["conn.errors"], 3)
	assert.EqualValues(t, stat["conn.closing"], 2)
	assert.EqualValues(t, stat["messages.total"], 9801)
	assert.EqualValues(t, stat["messages.errors"], 42)
}
