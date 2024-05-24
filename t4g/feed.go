package t4g

import (
	"context"
	"fmt"
	"strconv"
	"time"

	mapset "github.com/deckarep/golang-set/v2"
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

// AddNewEventsToFeed will add new events in the passed events to the feed
func AddNewEventsToFeed(feed *feeds.Feed, events []Event) *feeds.Feed {
	// Get feed ids, ignoring ones that cannot be converted to a number
	// Items without a number will be sorted to the end and eventually removed
	feedIds := mapset.NewSetWithSize[int](len(feed.Items))
	for _, item := range feed.Items {
		feedId, err := strconv.Atoi(item.Id)
		if err == nil {
			feedIds.Add(feedId)
		}
	}

	// Add events to feed that do not already exist
	for _, event := range events {
		if !feedIds.Contains(event.Id) {
			eventItem := event.ToFeedItem()
			feed.Add(eventItem)
		}
	}

	// Sort feed items
	feed.Sort(feedSortFunc)

	// Keep length of feed to 60 items (5 pages of 12 events)
	if len(feed.Items) > 60 {
		feed.Items = feed.Items[:60]
	}

	// Update the updated time
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

func feedSortFunc(item1, item2 *feeds.Item) bool {
	feedId1, err := strconv.Atoi(item1.Id)
	if err == nil {
		feedId1 = -1
	}

	feedId2, err := strconv.Atoi(item2.Id)
	if err == nil {
		feedId2 = -1
	}

	return feedId1 < feedId2
}
