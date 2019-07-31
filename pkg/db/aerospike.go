package db

import (
	"fmt"

	aero "github.com/aerospike/aerospike-client-go"
)

// AerospikeClient is a wrapper around the official package
type AerospikeClient struct {
	Client    *aero.Client
	Namespace string
}

// NewAerospikeClient connects and return an aerospike client instance where to store/read information
func NewAerospikeClient(addr, namespace string, port int) *AerospikeClient {
	client, err := aero.NewClient(addr, port)
	if err != nil {
		panic(err)
	}

	return &AerospikeClient{Client: client, Namespace: namespace}
}

// Record struct contains the result data of a query
type Record struct {
	Bins       map[string]interface{} `json:"bins"`
	Expiration uint32                 `json:"exp"`
}

// GetOne returns the associated Record (aka Bins) for the given key object
func (ac *AerospikeClient) GetOne(setName string, key interface{}) (interface{}, error) {
	k, err := aero.NewKey(ac.Namespace, setName, key)
	if err != nil {
		return nil, fmt.Errorf("could not create key: %v", err)
	}

	_, err = ac.Client.Exists(ac.Client.DefaultPolicy, k)
	if err != nil {
		return nil, fmt.Errorf("key does not exist: %v", err)
	}

	r, err := ac.Client.Get(ac.Client.DefaultPolicy, k)
	if err != nil {
		return nil, fmt.Errorf("could not get record: %v", err)
	}

	return &Record{
		Bins:       r.Bins,
		Expiration: r.Expiration,
	}, nil
}

// AddOne add the map value to the specified key in the set
func (ac *AerospikeClient) AddOne(setName string, key interface{}, value map[string]interface{}) error {
	k, err := aero.NewKey(ac.Namespace, setName, key)
	if err != nil {
		return fmt.Errorf("could not create key: %v", err)
	}

	err = ac.Client.Add(ac.Client.DefaultWritePolicy, k, value)
	if err != nil {
		return fmt.Errorf("could not add key/value pair: %v", err)
	}
	return nil
}

// Close will be called as defer from the dependency whenever it's needed to
// close the connection
func (ac *AerospikeClient) Close() error {
	ac.Client.Close()
	return nil
}

// Health will return a ping based on whether the database is healthy
func (ac *AerospikeClient) Health() error {
	if !ac.Client.IsConnected() {
		return fmt.Errorf("database is not connected")
	}
	return nil
}

// DeleteOne deletes a single record in the specified set
func (ac *AerospikeClient) DeleteOne(setName string, key interface{}) error {
	k, err := aero.NewKey(ac.Namespace, setName, key)
	if err != nil {
		return fmt.Errorf("could not create key: %v", err)
	}

	_, err = ac.Client.Exists(ac.Client.DefaultPolicy, k)
	if err != nil {
		return fmt.Errorf("key does not exist: %v", err)
	}

	_, err = ac.Client.Delete(ac.Client.DefaultWritePolicy, k)
	if err != nil {
		return fmt.Errorf("could not delete the record record: %v", err)
	}
	return nil
}

// TruncateSet will remove all the keys in a set asynchronously based on the time specified
// If time = nil
func (ac *AerospikeClient) TruncateSet(setName string) error {
	return ac.Client.Truncate(ac.Client.DefaultWritePolicy, ac.Namespace, setName, nil)
}
