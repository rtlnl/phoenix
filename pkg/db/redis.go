package db

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

// RedisClient is a wrapper around the official package
type RedisClient struct {
	Client *redis.ClusterClient
}

// NewRedisClient connects and return a redis client instance where to store/read information
func NewRedisClient(addrs []string, password string) *RedisClient {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Password: password,
		Addrs:    addrs,
	})

	// ReloadState reloads cluster state. It calls ClusterSlots func
	// to get cluster slots information.
	err := client.ReloadState()
	if err != nil {
		panic(err)
	}

	return &RedisClient{Client: client}
}

// BulkImport tries to insert all the personalized content items into the database
func (rc *RedisClient) BulkImport(u uuid.UUID, f io.ReadCloser) error {
	rd := bufio.NewReader(f)
	header := true
	for {
		l, err := rd.ReadString('\n')
		if err == io.EOF {
			break
		}

		// skip header
		if header {
			header = false
			continue
		}

		// Key for personalization => UUID:user:ID -> [ ... ]
		uID, rItems := splitRecommendationRecord(l)

		// TODO: format value/key here
		key := fmt.Sprintf("%s:user:%s", u.String(), uID)
		if o := rc.Client.Set(key, rItems, 0); o.Err() != nil {
			return o.Err()
		}
	}
	return nil
}

// splitRecommendationRecord splits the incoming record and return the userID and the
// recommended items for that user
func splitRecommendationRecord(record string) (string, string) {
	values := strings.Split(record, ",")
	uID := strings.Split(values[0], "_")[0]
	return uID, values[1]
}

// GetValue returns a value (or an error) given the key as input
func (rc *RedisClient) GetValue(key string) (string, error) {
	res := rc.Client.Get(key)
	return res.Result()
}

// SetValue sets a new entry in the database
func (rc *RedisClient) SetValue(key, value string) error {
	res := rc.Client.Set(key, value, 0)
	return res.Err()
}
