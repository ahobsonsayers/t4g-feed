package t4g

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/feeds"
	"github.com/samber/lo"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	mapset "github.com/deckarep/golang-set/v2"
)

const maxFeedItems = 60

type Feed struct {
	location *string
	feed     *feeds.Feed
	mutex    sync.Mutex
}

func (f *Feed) UpdatedAt() time.Time {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.feed.Updated
}

func (f *Feed) Update(ctx context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Get events
	events, err := Events(ctx, f.location, lo.ToPtr(5))
	if err != nil {
		return err
	}

	// Get current feed ids, ignoring ones that cannot be converted to a number
	// Items without a number will be sorted to the end and eventually removed
	feedIds := mapset.NewSetWithSize[int](len(f.feed.Items))
	var minFeedId int
	for _, item := range f.feed.Items {
		feedId, err := strconv.Atoi(item.Id)
		if err != nil {
	    continue
		}
		feedIds.Add(feedId)
		if minFeedId == 0 || feedId < minFeedId {
		  minFeedId = feedId
		}
	}

	// Add events to feed that do not already exist
	// and are larger than the current min feed id as
	// sometimes events are removed and older events
	// creep back in.
	for _, event := range events {
		if !feedIds.Contains(event.Id) && event.Id > minFeedId {
			feedItem := eventToFeedItem(event)
			f.feed.Add(feedItem)
		}
	}

	// Sort feed items
	f.feed.Sort(feedSortFunc)

	// Keep length of feed to 60 items (5 pages of 12 events)
	if len(f.feed.Items) > maxFeedItems {
		f.feed.Items = f.feed.Items[:maxFeedItems]
	}

	// Update the updated time
	f.feed.Updated = time.Now()

	return nil
}

func (f *Feed) ToRss() (string, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.feed.ToRss()
}

func NewFeed(location *string) *Feed {
	feedTitle := "T4G Feed"
	feedDescription := "Tickets For Good Events"
	if location != nil {
		titleLocation := cases.Title(language.English).String(*location)

		feedTitle = fmt.Sprintf("%s: %s", feedTitle, titleLocation)
		feedDescription = fmt.Sprintf("%s in %s", feedDescription, titleLocation)
	}

	return &Feed{
		location: location,
		feed: &feeds.Feed{
			Title:       feedTitle,
			Link:        &feeds.Link{Href: EventsUrl(EventsInput{Location: location})},
			Description: feedDescription,
		},
	}
}

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
		feed = NewFeed(location)
		cachedFeeds[lo.FromPtr(location)] = feed

		if len(cachedFeeds) > 10 {
			deleteOldestCachedFeed(cachedFeeds)
		}
	}

	// If there is a debounce and we are within the debounce period, return the cached feed
	if debounceTime != nil && time.Since(feed.UpdatedAt()) < *debounceTime {
		return feed, nil
	}

	// Update feed
	err := feed.Update(ctx)
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
