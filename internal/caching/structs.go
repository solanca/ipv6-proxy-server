package caching

import (
	"atlas/pkg/goccm"
	cache "github.com/patrickmn/go-cache"
	"sync"
	"time"
)

var (
	// Accounts is a slice of maps that contain the account information
	Accounts = cache.New(5*time.Minute, 10*time.Minute)

	// Concurrent is a map of concurrency managers
	Concurrent map[string]goccm.ConcurrencyManager

	mu = sync.RWMutex{}
)
