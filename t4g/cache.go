package t4g

import (
	"sync"
	"time"

	"github.com/gorilla/feeds"
)

var (
	cachedFeeds      = map[string]*feeds.Feed{} // Location -> feed
	cachedFeedsMutex sync.Mutex
)

func deleteOldestCachedFeed(cachedFeeds map[string]*feeds.Feed) {
	oldestTime := time.Now()

	// Find the oldest cached feed
	var oldestKey string
	for key, cachedFeed := range cachedFeeds {
		if cachedFeed.Updated.Before(oldestTime) {
			oldestKey = key
			oldestTime = cachedFeed.Updated
		}
	}

	delete(cachedFeeds, oldestKey)
}
