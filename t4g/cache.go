package t4g

import (
	"sync"
	"time"
)

var (
	cachedFeeds      = map[string]*Feed{} // Location -> feed
	cachedFeedsMutex sync.Mutex
)

func deleteOldestCachedFeed(cachedFeeds map[string]*Feed) {
	oldestTime := time.Now()

	// Find the oldest cached feed
	var oldestKey string
	for key, cachedFeed := range cachedFeeds {
		if cachedFeed.UpdatedAt().Before(oldestTime) {
			oldestKey = key
			oldestTime = cachedFeed.UpdatedAt()
		}
	}

	delete(cachedFeeds, oldestKey)
}
