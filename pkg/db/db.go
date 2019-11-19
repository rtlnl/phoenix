package db

// DB is the interface that will allow to use different backends
// for storing data into the database
type DB interface {
	GetOne(table string, key string) (string, error)
	AddOne(table string, key string, value string) error
	GetAllRecords(table string) (map[string]string, error)
	DeleteOne(table string, key string) error
	DropTable(table string) error
	Close() error
	Health() error
}
