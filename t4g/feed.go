package t4g

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/feeds"
	"github.com/samber/lo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func NewFeed(location *string) *feeds.Feed {
	feedTitle := "T4G Feed"
	feedDescription := "Tickets For Good Events"
	if location != nil {
		titleLocation := cases.Title(language.English).String(*location)

		feedTitle = fmt.Sprintf("%s: %s", feedTitle, titleLocation)
		feedDescription = fmt.Sprintf("%s in %s", feedDescription, titleLocation)
	}

	return &feeds.Feed{
		Title:       feedTitle,
		Link:        &feeds.Link{Href: EventsUrl(EventsInput{Location: location})},
		Description: feedDescription,
	}
}

// AddNewEventsToFeed will add new events in the passed event to the feed
func AddNewEventsToFeed(feed *feeds.Feed, events []Event) *feeds.Feed {
	// Get latest event id in the feed as a number.
	// If we fail to parse, latestEventId will be 0, meaning all passed
	// events will be deemed newer and will be added to the feed.
	var latestEventId int
	if len(feed.Items) > 0 {
		latestEventId, _ = strconv.Atoi(feed.Items[0].Id)
	}

	// Get events newer than the latest event in the feed i.e. has a larger event id.
	// If we fail to parse the event id, add event to feed regardless
	newFeedItems := make([]*feeds.Item, 0, len(events))
	for _, event := range events {
		eventId, err := strconv.Atoi(event.Id)
		if eventId > latestEventId || err != nil {
			newerFeedItem := event.ToFeedItem()
			newFeedItems = append(newFeedItems, newerFeedItem)
		}
	}

	// Get updated feed items by concatenating new feed items and
	// current feed items, only keeping the latest 100 items
	updateFeedItems := make([]*feeds.Item, 0, len(newFeedItems)+len(feed.Items))
	updateFeedItems = append(updateFeedItems, newFeedItems...)
	updateFeedItems = append(updateFeedItems, feed.Items...)
	if len(updateFeedItems) > 100 {
		updateFeedItems = feed.Items[:100]
	}

	// Update feed, updating time regardless of whether new items added
	feed.Items = updateFeedItems
	feed.Updated = time.Now()

	return feed
}

// FetchFeed will fetch a Tickets For Good events feed for a location.
// If debounce time is set and the time period has not passed since the time the
// feed was last updated and the function call, a cached feed will be returned.
// If the time period has pass, the feed will be fully retched.
func FetchFeed(ctx context.Context, location *string, debounceTime *time.Duration) (*feeds.Feed, error) {
	cachedFeedsMutex.Lock()
	defer cachedFeedsMutex.Unlock()

	// Get cached feed. If there is no cached feed, create a new one.
	// Only keep the 10 most recently updated feeds
	feed, isCached := cachedFeeds[lo.FromPtr(location)]
	if !isCached {
		feed = NewFeed(location)
		cachedFeeds[lo.FromPtr(location)] = feed

		if len(cachedFeeds) > 10 {
			deleteOldestCachedFeed(cachedFeeds)
		}
	}

	// If there is a debounce and we are within the debounce period, return the cached feed
	if debounceTime != nil && time.Since(feed.Updated) < *debounceTime {
		return feed, nil
	}

	// Get events
	events, err := Events(ctx, location, lo.ToPtr(5))
	if err != nil {
		return nil, err
	}

	// Add events to feed
	AddNewEventsToFeed(feed, events)

	return feed, nil
}
