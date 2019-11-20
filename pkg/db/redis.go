package db

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/rs/zerolog/log"
)

// Redis is a wrapper struct around Redis official package
type Redis struct {
	*redis.Client
}

// NewRedisClient creates and open a new connection to Redis
func NewRedisClient(addr string, opts ...func(*redis.Options)) (*Redis, error) {
	o := &redis.Options{
		Addr: addr,
	}

	// call option functions on instance to set options on it
	for _, opt := range opts {
		opt(o)
	}

	// construct client
	client := redis.NewClient(o)

	i, err := client.Ping().Result()
	if err != nil && i != "PONG" {
		return nil, err
	}

	return &Redis{client}, nil
}

// Password functional option
func Password(p string) func(*redis.Options) {
	return func(r *redis.Options) {
		r.Password = p
	}
}

// Close will be called as defer from the dependency whenever it's needed to
// close the connection
func (db *Redis) Close() error {
	return db.Client.Close()
}

// Health will return a ping based on whether the database is healthy
func (db *Redis) Health() error {
	return db.Ping().Err()
}

// GetOne returns the value associated with that key
func (db *Redis) GetOne(table, key string) (string, error) {
	val, err := db.Client.HGet(table, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key %s not found", key)
	} else if err != nil {
		return "", err
	}
	return val, nil
}

// AddOne store the key/value in the redis
func (db *Redis) AddOne(table, key string, values string) error {
	return db.Client.HSet(table, key, values).Err()
}

// DeleteOne deletes a key from a table
func (db *Redis) DeleteOne(table, key string) error {
	err := db.Client.HDel(table, key).Err()
	if err == redis.Nil {
		return fmt.Errorf("table %s not found", table)
	} else if err != nil {
		return err
	}
	return nil
}

// DropTable deletes all the keys and the table itself
func (db *Redis) DropTable(table string) error {
	err := db.Client.Del(table).Err()
	if err == redis.Nil {
		return fmt.Errorf("table %s not found", table)
	} else if err != nil {
		return err
	}
	return nil
}

// GetAllRecords returns all the records from that table
// the map[string]string represents the signalID -> recommendations encoded
func (db *Redis) GetAllRecords(table string) (map[string]string, error) {
	elems := map[string]string{}
	iter := db.Client.HScan(table, 0, "*", maxScan).Iterator()

	counter := 0
	key := ""
	for iter.Next() {
		// stop iterating
		if counter == maxEntries + 1 {
			return elems, nil
		}
		if iter.Err() != nil {
			log.Error().Msgf("error found: %s",iter.Err().Error())
			continue
		}
		val := iter.Val()
		if counter % 2 == 0 {
			key = val
		} else {
			elems[key] = val
		}
		counter++
	}
	return elems, nil
}
