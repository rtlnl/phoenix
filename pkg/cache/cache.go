package cache

import "github.com/rtlnl/phoenix/models"

/*

# How to store in the cache.

Since the Key needs to be unique, we need to find a way to store the actual information in a unique manner.

Phoenix has the concept of models, which by definition, it needs to be unique. This is a good starting point. Since the
recommendation system is based on a signal_ID to retrieve the recommended items, the storing strategy in the cache layer should be
the following

Key:   modelName#signalID
Value: [{...}, {...}, ... ]

## Benchmarks:

BenchmarkWriteToCache/1-shards-12         	  287896	      3729 ns/op	   10231 B/op	     185 allocs/op
BenchmarkWriteToCache/256-shards-12       	  307555	      3347 ns/op	   10242 B/op	     185 allocs/op
BenchmarkWriteToCache/512-shards-12       	  367656	      3227 ns/op	   10239 B/op	     185 allocs/op
BenchmarkWriteToCache/1024-shards-12      	  320841	      3379 ns/op	   10285 B/op	     185 allocs/op
BenchmarkWriteToCache/2048-shards-12      	  342153	      3451 ns/op	   10319 B/op	     185 allocs/op
BenchmarkWriteToCache/4096-shards-12      	  327483	      3529 ns/op	   10309 B/op	     185 allocs/op
BenchmarkWriteToCache/8192-shards-12      	  340054	      3581 ns/op	   10497 B/op	     185 allocs/op
*/

// Cache is the interface that will be used to create a caching layer to speedup the
// serving of the recommendations
type Cache interface {
	Set(key string, value []models.ItemScore) bool
	Get(key string) ([]models.ItemScore, bool)
	Del(key string) bool
	Empty() bool
}
