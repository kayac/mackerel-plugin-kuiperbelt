package kuiperbelt

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	mp "github.com/mackerelio/go-mackerel-plugin"
)

// Plugin mackerel plugin for kuiperbelt
type Plugin struct {
	Target   string
	Tempfile string
	Prefix   string
}

// MetricKeyPrefix interface for PluginWithPrefix
func (m Plugin) MetricKeyPrefix() string {
	if m.Prefix == "" {
		m.Prefix = "kuiperbelt"
	}
	return m.Prefix
}

// FetchMetrics interface for mackerelplugin
func (m Plugin) FetchMetrics() (map[string]float64, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/stats", m.Target))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	stats := struct {
		Connections        float64 `json:"connections"`
		TotalConnections   float64 `json:"total_connections"`
		TotalMessages      float64 `json:"total_messages"`
		ConnectErrors      float64 `json:"connect_errors"`
		MessageErrors      float64 `json:"message_errors"`
		ClosingConnections float64 `json:"closing_connections"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}
	ret := make(map[string]float64, 6)
	ret["conn_current"] = stats.Connections
	ret["conn_total"] = stats.TotalConnections
	ret["conn_errors"] = stats.ConnectErrors
	ret["conn_closing"] = stats.ClosingConnections
	ret["messages_total"] = stats.TotalMessages
	ret["messages_errors"] = stats.MessageErrors

	return ret, nil
}

// GraphDefinition interface for mackerelplugin
func (m Plugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := strings.Title(m.Prefix)

	// https://github.com/kayac/go-kuiperbelt#stats
	var graphdef = map[string]mp.Graphs{
		"conn": {
			Label: (labelPrefix + " Connections"),
			Unit:  mp.UnitInteger,
			Metrics: []mp.Metrics{
				{Name: "conn_current", Label: "Current", Diff: false},
				{Name: "conn_closing", Label: "Closing", Diff: false},
				{Name: "conn_total", Label: "New", Diff: true},
				{Name: "conn_errors", Label: "Errors", Diff: true},
			},
		},
		"messages": {
			Label: (labelPrefix + " Messages"),
			Unit:  mp.UnitInteger,
			Metrics: []mp.Metrics{
				{Name: "messages_total", Label: "Messages", Diff: true},
				{Name: "messages_errors", Label: "Errors", Diff: true},
			},
		},
	}
	return graphdef
}

// Do the plugin
func Do() {
	optHost := flag.String("host", "localhost", "Hostname")
	optPort := flag.String("port", "9180", "Port")
	optPrefix := flag.String("metric-key-prefix", "kuiperbelt", "Metric key prefix")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	flag.Parse()

	var kuiperbelt Plugin
	kuiperbelt.Prefix = *optPrefix
	kuiperbelt.Target = fmt.Sprintf("%s:%s", *optHost, *optPort)

	helper := mp.NewMackerelPlugin(kuiperbelt)
	helper.Tempfile = *optTempfile
	helper.Run()
}
