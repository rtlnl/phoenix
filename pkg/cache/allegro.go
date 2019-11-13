package cache

import (
	"encoding/json"
	"time"

	"github.com/allegro/bigcache"
	"github.com/rs/zerolog/log"
	"github.com/rtlnl/phoenix/models"
)

// AllegroBigCache is the struct holding the Cache layer object
type AllegroBigCache struct {
	*bigcache.BigCache
}

// Shards functional option
func Shards(s int) func(*bigcache.Config) {
	return func(c *bigcache.Config) {
		c.Shards = s
	}
}

// Verbose functional option
func Verbose(v bool) func(*bigcache.Config) {
	return func(c *bigcache.Config) {
		c.Verbose = v
	}
}

// LifeWindow functional option
func LifeWindow(l time.Duration) func(*bigcache.Config) {
	return func(c *bigcache.Config) {
		c.LifeWindow = l
	}
}

// MaxEntriesInWindow functional option
func MaxEntriesInWindow(m int) func(*bigcache.Config) {
	return func(c *bigcache.Config) {
		c.MaxEntriesInWindow = m
	}
}

// MaxEntrySize functional option
func MaxEntrySize(m int) func(*bigcache.Config) {
	return func(c *bigcache.Config) {
		c.MaxEntrySize = m
	}
}

// NewAllegroBigCache returns a new AllegroCache object
func NewAllegroBigCache(opts ...func(*bigcache.Config)) (*AllegroBigCache, error) {
	c := &bigcache.Config{}

	// call option functions on instance to set options on it
	for _, opt := range opts {
		opt(c)
	}

	cache, err := bigcache.NewBigCache(*c)
	if err != nil {
		return nil, err
	}
	return &AllegroBigCache{cache}, nil
}

// Set stores a key/value pair with specified weight into the cache layer
func (ac *AllegroBigCache) Set(key string, value []models.ItemScore) bool {
	v, err := json.Marshal(value)
	if err != nil {
		return false
	}
	if err := ac.BigCache.Set(key, []byte(v)); err != nil {
		return false
	}
	return true
}

// Get returns the value associated with the particular key
func (ac *AllegroBigCache) Get(key string) ([]models.ItemScore, bool) {
	v, err := ac.BigCache.Get(key)
	if err != nil {
		return nil, false
	}

	var value []models.ItemScore
	if err := json.Unmarshal(v, &value); err != nil {
		log.Error().Msgf("key %s not unmarshalled correctly", key)
		return nil, false
	}
	return value, true
}

// Del deletes the entry from the cache layer
func (ac *AllegroBigCache) Del(key string) bool {
	if err := ac.BigCache.Delete(key); err != nil {
		return false
	}
	return true
}

// Empty clears the cache
func (ac *AllegroBigCache) Empty() bool {
	if err := ac.BigCache.Reset(); err != nil {
		return false
	}
	return true
}

// Close closes the cache once it is not necessary anymore
func (ac *AllegroBigCache) Close() {
	if err := ac.BigCache.Close(); err != nil {
		log.Error().Msgf("could not close the cache %v", err)
	}
}
