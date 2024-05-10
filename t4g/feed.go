package t4g

import (
	"context"
	"sync"
	"time"

	"github.com/gorilla/feeds"
)

func Feed(ctx context.Context, location *string) (*feeds.Feed, error) {
	events, err := Events(ctx, location)
	if err != nil {
		return nil, err
	}

	// Create feed
	feed := &feeds.Feed{
		Title:       "Tickets For Good Feed",
		Link:        &feeds.Link{Href: EventsUrl(EventsInput{Location: location})},
		Description: "Tickets For Good events in London",
		Updated:     time.Now(),
	}
	for _, event := range events {
		feed.Add(event.ToFeedItem())
	}

	return feed, nil
}

type cachedFeed struct {
	feed    *feeds.Feed
	updated time.Time
}

var (
	cachedFeeds      = map[string]cachedFeed{} // Location -> Cached feed
	cachedFeedsMutex sync.Mutex
)

// FeedDebounce is the same as feed, but will only re get feed if
// debounce time has passed since last call. Otherwise cached feed
// will be returned
func FeedDebounce(ctx context.Context, location *string, debounceTime time.Duration) (*feeds.Feed, error) {
	var locationString string
	if location != nil {
		locationString = *location
	}

	cachedFeedsMutex.Lock()
	defer cachedFeedsMutex.Unlock()

	// Get cached feed
	// Return cached feed if we are within the debounce period
	feed, ok := cachedFeeds[locationString]
	if ok && time.Since(feed.updated) < debounceTime {
		return feed.feed, nil
	}

	// If getting new feed, keep number of cached feeds to 10
	if !ok && len(cachedFeeds) > 10 {
		deleteOldestCachedFeed(cachedFeeds)
	}

	// Get new feed
	newFeed, err := Feed(ctx, location)
	if err != nil {
		return nil, err
	}

	// Update cached feeds
	cachedFeeds[locationString] = cachedFeed{
		feed:    newFeed,
		updated: time.Now(),
	}

	return newFeed, nil
}

func deleteOldestCachedFeed(cachedFeeds map[string]cachedFeed) {
	oldestTime := time.Now()

	// Find the oldest cached feed
	var oldestKey string
	for key, cachedFeed := range cachedFeeds {
		if cachedFeed.updated.Before(oldestTime) {
			oldestKey = key
			oldestTime = cachedFeed.updated
		}
	}

	delete(cachedFeeds, oldestKey)
}
