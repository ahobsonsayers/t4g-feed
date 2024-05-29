package t4g

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/feeds"
	"github.com/samber/lo"
)

const (
	maxFeedItems  = 75
	numEventPages = 5 // Number of event pages to get on update
)

var (
	cachedFeeds      = map[string]*Feed{} // Location -> feed
	cachedFeedsMutex sync.Mutex
)

// FetchFeed will fetch a Tickets For Good events feed for a location.
// If debounce time is set and the time period has not passed since the time the
// feed was last updated and the function call, a cached feed will be returned.
// If the time period has pass, the feed will be fully retched.
func FetchFeed(ctx context.Context, location *string, debounceTime *time.Duration) (*Feed, error) {
	cachedFeedsMutex.Lock()
	defer cachedFeedsMutex.Unlock()

	// Get cached feed. If there is no cached feed, create a new one.
	// Only keep the 10 most recently updated feeds
	feed, isCached := cachedFeeds[lo.FromPtr(location)]
	if !isCached {
		feed = NewFeed(location, lo.ToPtr(maxFeedItems))
		cachedFeeds[lo.FromPtr(location)] = feed

		if len(cachedFeeds) > 10 {
			deleteOldestCachedFeed(cachedFeeds)
		}
	}

	// If there is a debounce and we are within the debounce period, return the cached feed
	if debounceTime != nil && time.Since(feed.UpdatedAt()) < *debounceTime {
		return feed, nil
	}

	// Update feed with event pages
	err := feed.Update(ctx, numEventPages)
	if err != nil {
		return nil, err
	}

	return feed, nil
}

func eventToFeedItem(event Event) *feeds.Item {
	return &feeds.Item{
		Id:          lo.Ternary(event.Id == 0, "", strconv.Itoa(event.Id)),
		Title:       event.Title,
		Link:        &feeds.Link{Href: event.Link},
		Description: fmt.Sprintf("%s | %s | %s", event.Date, event.Location, event.Category),
		Enclosure:   &feeds.Enclosure{Url: event.Image, Type: "image/jpeg", Length: "0"},
		Created:     time.Now(),
	}
}

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
