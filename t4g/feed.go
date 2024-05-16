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
	// We only want to add events that do not already exist.
	// - Generally new events will have a larger id than the largest id in the feed.
	// - Sometimes new events are inserted in the middle, and do not have a larger id.
	// - If an event is removed, older events might sneak in at the end. These are not new and should not be included!
	// Therefore we will include all new events, with id LARGER than the earliest id in the feed.

	feedIds := mapset.NewSetWithSize[int](len(feed.Items))
	var minFeedId int
	for _, item := range feed.Items {
		feedId, err := strconv.Atoi(item.Id)
		if err == nil {
			feedIds.Add(feedId)
			if feedId < minFeedId {
				minFeedId = feedId
			}
		}
	}

	eventIds := mapset.NewSetWithSize[int](len(events))
	for _, event := range events {
		eventIds.Add(event.Id)
	}

	newEventIds := eventIds.Difference(feedIds)

	newFeedItems := make([]*feeds.Item, 0, len(events))
	for _, event := range events {
		if event.Id == 0 || (newEventIds.Contains(event.Id) && event.Id > minFeedId) {
			newFeedItems = append(newFeedItems, event.ToFeedItem())
		}
	}

	// Get updated feed items by concatenating new feed items and
	// current feed items, only keeping the latest 100 items
	updatedFeedItems := make([]*feeds.Item, 0, len(newFeedItems)+len(feed.Items))
	updatedFeedItems = append(updatedFeedItems, newFeedItems...)
	updatedFeedItems = append(updatedFeedItems, feed.Items...)
	if len(updatedFeedItems) > 100 {
		updatedFeedItems = feed.Items[:100]
	}

	// Update feed, updating time regardless of whether new items added
	feed.Items = updatedFeedItems
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
