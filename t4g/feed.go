package t4g

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/feeds"
	"github.com/samber/lo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	mapset "github.com/deckarep/golang-set/v2"
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

func AddEventsToFeed(feed *feeds.Feed, events []Event) *feeds.Feed {
	feedItemIds := mapset.NewSetWithSize[string](len(feed.Items))
	for _, feedItem := range feed.Items {
		feedItemIds.Add(feedItem.Id)
	}

	eventIds := mapset.NewSetWithSize[string](len(events))
	for _, event := range events {
		eventIds.Add(event.Id)
	}

	// Determine the added events (i.e. not already in the feed) and add them.
	// Treat events without an id as added.
	addedEventIds := eventIds.Difference(feedItemIds)
	for _, event := range events {
		if event.Id == "" || addedEventIds.Contains(event.Id) {
			feed.Add(event.ToFeedItem())
		}
	}

	// Only keep last 100 added items
	if len(feed.Items) > 100 {
		feed.Items = feed.Items[:100]
	}

	// Update feed updated time regardless of whether new items added
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
	AddEventsToFeed(feed, events)

	return feed, nil
}
