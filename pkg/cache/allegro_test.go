package cache

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/rtlnl/phoenix/models"
	"github.com/stretchr/testify/assert"
)

func TestNewAllegroCache(t *testing.T) {
	a, err := NewAllegroBigCache()
	if err != nil {
		t.Fail()
	}
	assert.NotNil(t, a)
}

func createAllegroBigCache() *AllegroBigCache {
	ac, err := NewAllegroBigCache(Shards(1024),
		LifeWindow(time.Minute*10),
		MaxEntriesInWindow(1000*10*60),
		MaxEntrySize(500),
	)
	if err != nil {
		panic(err)
	}
	return ac
}

// This test is designed to make sure that the cache will
// not fail in case we run an Empty operation and we have write/read
// happening concurrently
func TestConcurrentEmpty(t *testing.T) {
	// create cache
	ac := createAllegroBigCache()
	defer ac.Close()
	defer ac.Empty()

	// fill the cache
	i := 1
	for i <= 1000000 {
		is := []models.ItemScore{
			{
				"score": "0.5",
				"type":  "movie",
				"item":  strconv.Itoa(i),
			},
		}
		if ok := ac.Set(strconv.Itoa(i), is); !ok {
			log.Info().Msg("error in setting the key")
			t.FailNow()
		}
		i++
	}

	var wg sync.WaitGroup

	for j := 1; j <= 1000; j++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// clear cache
			if index == 50 {
				if o := ac.Empty(); !o {
					log.Info().Msg("error in emptying the cache")
					t.Fail()
				}
			} else {
				// read values
				ac.Get(strconv.Itoa(index))

				// write values
				v := strconv.Itoa(i + 20)
				is := []models.ItemScore{
					{
						"score": "0.5",
						"type":  "movie",
						"item":  v,
					},
				}
				ac.Set(v, is)
			}
		}(j)
	}

	wg.Wait()
}

func TestEmpty(t *testing.T) {
	// create cache
	ac := createAllegroBigCache()
	defer ac.Close()

	// fill the cache
	i := 1
	for i <= 10000 {
		v := strconv.Itoa(i)
		is := []models.ItemScore{
			{
				"score": "0.5",
				"type":  "movie",
				"item":  v,
			},
		}
		if ok := ac.Set(v, is); !ok {
			t.FailNow()
		}
		i++
	}

	// clear cache
	if ok := ac.Empty(); !ok {
		t.Fail()
	}

	assert.Equal(t, 0, int(ac.BigCache.Len()))
	assert.Equal(t, 0, int(ac.BigCache.Stats().Hits))
}

func TestSet(t *testing.T) {
	c := createAllegroBigCache()
	defer c.Empty()
	defer c.Close()

	is := []models.ItemScore{
		{
			"score": "0.5",
			"type":  "movie",
			"item":  "42",
		},
	}
	ok := c.Set("hello", is)

	assert.Equal(t, true, ok)
	assert.Equal(t, 1, int(c.BigCache.Len()))
}

func TestGet(t *testing.T) {
	c := createAllegroBigCache()
	defer c.Empty()
	defer c.Close()

	is := []models.ItemScore{
		{
			"score": "0.5",
			"type":  "movie",
			"item":  "42",
		},
	}
	if ok := c.Set("hello", is); !ok {
		t.Fail()
	}

	val, found := c.Get("hello")

	assert.Equal(t, true, found)
	assert.Equal(t, is, val)
}

func TestDel(t *testing.T) {
	c := createAllegroBigCache()
	defer c.Empty()
	defer c.Close()

	is := []models.ItemScore{
		{
			"score": "0.5",
			"type":  "movie",
			"item":  "42",
		},
	}
	if ok := c.Set("hello", is); !ok {
		t.Fail()
	}

	if ok := c.Del("hello"); !ok {
		t.Fail()
	}

	assert.Equal(t, 0, int(c.BigCache.Len()))
}

func BenchmarkWriteToCache(b *testing.B) {
	for _, shards := range []int{1, 256, 512, 1024, 2048, 4096, 8192} {
		b.Run(fmt.Sprintf("%d-shards", shards), func(b *testing.B) {
			writeToCache(b, shards, 100*time.Second, b.N)
		})
	}
}

func writeToCache(b *testing.B, shards int, lifeWindow time.Duration, requestsInLifeWindow int) {
	ac, _ := NewAllegroBigCache(Shards(shards), LifeWindow(lifeWindow), MaxEntriesInWindow(max(requestsInLifeWindow, 100)), MaxEntrySize(500))
	rand.Seed(time.Now().Unix())

	b.RunParallel(func(pb *testing.PB) {
		id := rand.Int()
		counter := 0

		b.ReportAllocs()
		for pb.Next() {
			ac.Set(fmt.Sprintf("key-%d-%d", id, counter), itemsBench)
			counter = counter + 1
		}
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// used to test the cache when benchmarking
var itemsBench = []models.ItemScore{
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
	{
		"score": "0.5",
		"type":  "movie",
		"item":  "42",
	},
}
