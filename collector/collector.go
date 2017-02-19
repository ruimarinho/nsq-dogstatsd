package collector

import (
	"encoding/json"
	"fmt"

	"github.com/ruimarinho/nsq-dogstatsd/internal/fetcher"
	"github.com/ruimarinho/nsq-dogstatsd/producer"
)

// Nodes wraps /nodes data.
type Nodes struct {
	StatusCode int       `json:"status_code"`
	StatusTxt  string    `json:"status_txt"`
	Data       NodesData `json:"data"`
}

// NodesData is an embedded Nodes type.
type NodesData struct {
	Producers []producer.Producer `json:"producers"`
}

// NSQDCollector holds a fetcher instance.
type NSQDCollector struct {
	Fetcher fetcher.Fetcher
}

// GetFetcher returns the fetcher instance of the collector.
func (nc NSQDCollector) GetFetcher() fetcher.Fetcher {
	return nc.Fetcher
}

// Info wraps `/info` data.
type Info struct {
	StatusCode int               `json:"status_code"`
	StatusTxt  string            `json:"status_txt"`
	Producer   producer.Producer `json:"data"`
}

// GetInfo retrieves and parses data from the /info endpoint of a nsqlookupd.
func (nc NSQDCollector) GetInfo() (Info, error) {
	var info Info

	body, err := nc.GetFetcher().Fetch("info")
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(body, &info)
	if err != nil {
		return info, err
	}

	if info.StatusCode != 200 {
		return info, fmt.Errorf("response code was %d", info.StatusCode)
	}

	return info, nil
}

// GetNodes retrieves and parses data from the /nodes endpoint of a nsqlookupd.
func (nc NSQDCollector) GetNodes() (Nodes, error) {
	var nodes Nodes

	body, err := nc.GetFetcher().Fetch("nodes")
	if err != nil {
		return nodes, err
	}

	err = json.Unmarshal(body, &nodes)
	if err != nil {
		return nodes, err
	}

	if nodes.StatusCode != 200 {
		return nodes, fmt.Errorf("response code was %d", nodes.StatusCode)
	}

	return nodes, nil
}
