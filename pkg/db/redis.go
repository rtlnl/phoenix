package db

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/rs/zerolog/log"
)

const (
	// LockOn sets the lock
	LockOn = "locked"
	// LockOff unsets the lock
	LockOff = "unlocked"
	// TTL for the lock
	TTL = 30 * time.Second
	// TTLRefreshInterval intervals for refreshing the TTL of the lock
	TTLRefreshInterval = 10 * time.Second
)

// Redis is a wrapper struct around Redis official package
type Redis struct {
	*redis.Client
	Pipeliner redis.Pipeliner
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
		log.Error().Err(err).Msg("REDIS could not get client")
		return nil, err
	}

	return &Redis{client, client.TxPipeline()}, nil
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
	ok, err := db.Client.HExists(table, key).Result()
	if !ok || err == redis.Nil || err != nil {
		return "", fmt.Errorf("key %s not found", key)
	}
	return db.Client.HGet(table, key).Result()
}

// AddOne store the key/value in the redis
func (db *Redis) AddOne(table, key string, values string) error {
	return db.Client.HSet(table, key, values).Err()
}

// DeleteOne deletes a key from a table
func (db *Redis) DeleteOne(table, key string) error {
	ok, err := db.Client.HExists(table, key).Result()
	if !ok || err == redis.Nil || err != nil {
		return fmt.Errorf("key %s not found", key)
	}
	return db.Client.HDel(table, key).Err()
}

// DropTable deletes all the keys and the table itself
func (db *Redis) DropTable(table string) error {
	_, err := db.Client.Exists(table).Result()
	if err == redis.Nil || err != nil {
		return fmt.Errorf("key %s not found", table)
	}
	return db.Client.Del(table).Err()
}

// GetAllRecords returns all the records from that table
// the map[string]string represents the signalID -> recommendations encoded
func (db *Redis) GetAllRecords(table string) (map[string]string, int, error) {
	elems := map[string]string{}
	iter := db.Client.HScan(table, 0, "*", maxScan).Iterator()

	counter := 0
	key := ""
	for iter.Next() {
		// stop iterating
		if counter == maxEntries+1 {
			return elems, -1, nil
		}
		if iter.Err() != nil {
			log.Error().Err(iter.Err()).Msg("REDIS could not get records")
			continue
		}
		val := iter.Val()
		if counter%2 == 0 {
			key = val
		} else {
			elems[key] = val
		}
		counter++
	}

	// retrieve number of elements
	count, err := db.Client.HLen(table).Result()
	if err == redis.Nil || err != nil {
		log.Error().Err(err).Msg("REDIS could not get count")
	}

	return elems, int(count), nil
}

// PipelineAddOne queues the HSET operation to the pipeline
func (db *Redis) PipelineAddOne(table, key string, values string) {
	db.Pipeliner.HSet(table, key, values)
}

// PipelineExec executes the commands in the Pipeline
func (db *Redis) PipelineExec() error {
	// we care only about the error and not the actual result of the command
	_, err := db.Pipeliner.Exec()
	return err
}

// Lock allows to lock the resource
func (db *Redis) Lock(key string) (bool, error) {
	res, err := db.Client.SetNX(key, LockOn, TTL).Result()
	if err == redis.Nil || err != nil || res == false {
		log.Error().Err(err).Msg("REDIS could not set key")
		return false, err
	}
	// the lock is on
	return true, nil
}

// Unlock unlocks the the key
func (db *Redis) Unlock(key string) (bool, error) {
	if err := db.Client.Del(key).Err(); err != nil {
		log.Error().Err(err).Msg("REDIS could not set key")
		return false, err
	}
	// now it's unlocked
	return true, nil
}

// ExtendTTL ddd
func (db *Redis) ExtendTTL(key string) error {
	err := db.Expire(key, TTL).Err()
	if err != nil && err == redis.Nil {
		return errors.New("REDIS lock has gone")
	}
	return nil
}
