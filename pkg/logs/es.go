package logs

import (
	"errors"

	es "github.com/elastic/go-elasticsearch/v7"
)

// ElasticSearchLogs is the struct that contains the information of ElasticSearch
type ElasticSearchLogs struct {
	Client    *es.Client
	IndexName string
}

// ESCredentials functional option for the kafka configuration
func ESCredentials(username, password string) func(*es.Config) {
	return func(cfg *es.Config) {
		cfg.Username = username
		cfg.Password = password
	}
}

// NewElasticSearchLogs creates a new object for sending logs to ElasticSearch
// For functional options see:
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
func NewElasticSearchLogs(addresses []string, index string, options ...func(*es.Config)) (ElasticSearchLogs, error) {
	cfg := &es.Config{
		Addresses: addresses,
	}

	// call option functions on instance to set options on it
	for _, opt := range options {
		opt(cfg)
	}

	es, err := es.NewClient(*cfg)
	if err != nil {
		return ElasticSearchLogs{}, err
	}

	return ElasticSearchLogs{
		Client: es,
	}, nil
}

func (es *ElasticSearchLogs) Write(rl RowLog) error {
	return errors.New("not implemented yet")
}
