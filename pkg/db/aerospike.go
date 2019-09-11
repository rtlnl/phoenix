package db

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	aero "github.com/aerospike/aerospike-client-go"
)

// AerospikeClient is a wrapper around the official package
type AerospikeClient struct {
	Client    *aero.Client
	Namespace string

	basePolicy    *aero.BasePolicy  // Base policy for reading data
	scanPolicy    *aero.ScanPolicy  // Special Policy for scanning data
	writingPolicy *aero.WritePolicy // Special policy for writing data
}

// NewAerospikeClient connects and return an aerospike client instance where to store/read information
func NewAerospikeClient(addr, namespace string, port int) *AerospikeClient {
	client, err := aero.NewClient(addr, port)
	if err != nil {
		panic(err)
	}

	return &AerospikeClient{
		Client:        client,
		Namespace:     namespace,
		basePolicy:    aero.NewPolicy(),
		scanPolicy:    createNewScanPolicy(),
		writingPolicy: createNewWritingPolicy(),
	}
}

// Record struct contains the result data of a query
type Record struct {
	Key        string                 `json:"key"`
	Bins       map[string]interface{} `json:"bins"`
	Expiration uint32                 `json:"exp"`
}

// GetOne returns the associated Record (aka Bins) for the given key object
func (ac *AerospikeClient) GetOne(setName string, key string) (*Record, error) {
	k, err := aero.NewKey(ac.Namespace, setName, key)
	if err != nil {
		return nil, fmt.Errorf("could not create key %s. error %v", key, err)
	}

	e, err := ac.Client.Exists(ac.basePolicy, k)
	if err != nil {
		return nil, fmt.Errorf("key %s does not exist. error %v", key, err)
	}

	// key doesn't exists
	if e == false {
		return nil, fmt.Errorf("key %s does not exist", key)
	}

	r, err := ac.Client.Get(ac.basePolicy, k)
	if err != nil {
		return nil, fmt.Errorf("could not get record. error %v", err)
	}

	return &Record{
		Key:        r.Key.Value().String(),
		Bins:       r.Bins,
		Expiration: r.Expiration,
	}, nil
}

// AddOne add the map value to the specified key in the set
func (ac *AerospikeClient) AddOne(setName string, key string, binKey string, binValue interface{}) error {
	k, err := aero.NewKey(ac.Namespace, setName, key)
	if err != nil {
		return fmt.Errorf("could not create key: %v", err)
	}

	bin := aero.NewBin(binKey, binValue)

	err = ac.Client.PutBins(ac.writingPolicy, k, bin)
	if err != nil {
		return fmt.Errorf("could not add key/value pair: %v", err)
	}
	return nil
}

// AddRecord is used to add an already wrapped object into Aerospike database
func (ac *AerospikeClient) AddRecord(setName string, key aero.Value, value aero.BinMap) error {
	// value can contain multiple bins object (aka map[string]interface{})
	// hence we need to iterate through it
	for k, v := range value {
		newKey, err := aero.NewKey(ac.Namespace, setName, key.String())
		if err != nil {
			return fmt.Errorf("could not add key/value pair: %v", err)
		}

		b := aero.NewBin(k, v)
		if err := ac.Client.PutBins(ac.writingPolicy, newKey, b); err != nil {
			return fmt.Errorf("could not add key/value pair: %v", err)
		}
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
func (ac *AerospikeClient) DeleteOne(setName string, key string) error {
	k, err := aero.NewKey(ac.Namespace, setName, key)
	if err != nil {
		return fmt.Errorf("could not create key: %v", err)
	}

	_, err = ac.Client.Exists(ac.basePolicy, k)
	if err != nil {
		return fmt.Errorf("key does not exist: %v", err)
	}

	_, err = ac.Client.Delete(ac.writingPolicy, k)
	if err != nil {
		return fmt.Errorf("could not delete the record record: %v", err)
	}
	return nil
}

// TruncateSet will remove all the keys in a set asynchronously based on the time specified
// If time = nil
func (ac *AerospikeClient) TruncateSet(setName string) error {
	return ac.Client.Truncate(ac.writingPolicy, ac.Namespace, setName, nil)
}

// GetAllRecords returns all the records of a specific set
func (ac *AerospikeClient) GetAllRecords(setName string) (*aero.Recordset, error) {
	return ac.Client.ScanAll(ac.scanPolicy, ac.Namespace, setName)
}

// AddMultipleRecords add an x amount of records to a specific set
func (ac *AerospikeClient) AddMultipleRecords(setName string, records *aero.Recordset) error {

	// buffer records for batch swapping
	recordsBuff := make(chan *aero.Result, 25000)

	var wg sync.WaitGroup
	wg.Add(runtime.NumCPU())

	defer wg.Wait()
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(records <-chan *aero.Result) {
			defer wg.Done()
			for res := range records {
				if err := ac.AddRecord(setName, res.Record.Key.Value(), res.Record.Bins); err != nil {
					log.Print(err)
				}
			}
		}(recordsBuff)
	}

	for res := range records.Results() {
		recordsBuff <- res
	}

	close(recordsBuff)

	// TODO: use channels/goroutines to improve the insert
	// for res := range records.Results() {
	// 	if err := ac.AddRecord(setName, res.Record.Key.Value(), res.Record.Bins); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

// Custom policy for scanning and reading the data in aerospike
func createNewScanPolicy() *aero.ScanPolicy {
	sp := aero.NewScanPolicy()
	sp.ConcurrentNodes = true
	sp.Priority = aero.LOW
	sp.IncludeBinData = true

	return sp
}

// Custom policy for writing/deleting data in aerospike
func createNewWritingPolicy() *aero.WritePolicy {
	wp := aero.NewWritePolicy(0, 0)
	wp.SendKey = true
	wp.RecordExistsAction = aero.UPDATE

	return wp
}
