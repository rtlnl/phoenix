package db

// DB is the interface that will allow to use different backends
// for storing data into the database
type DB interface {
	GetOne(table string, key string) (string, error)
	AddOne(table string, key string, value string) error
	GetAllRecords(table string) (map[string]string, error)
	DeleteOne(table string, key string) error
	DropTable(table string) error
	PipelineAddOne(table, key string, values string)
	PipelineExec() error
	Close() error
	Health() error
}

const (
	// maximum number of entries when previewing the data. Since Redis returns the key on the first iteration
	// then the value on the second one, and so on, we need to make sure that if we want to have 'x' amount
	// of complete k/v pairs, we need to double the amount of 'x'. For example, we want to have 25 complete k/v entries
	// hence maxEntries = 50
	maxEntries = 50
	// maximum number of elements return per scan in Redis. A large amount of Scan elements values, benefits the database
	// query system
	maxScan = 1000
)