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

type Feed struct {
	location *string
	maxItems int
	feed     *feeds.Feed
	mutex    sync.Mutex
}

func (f *Feed) UpdatedAt() time.Time {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.feed.Updated
}

func (f *Feed) Update(ctx context.Context, numEventPages int) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Get events
	events, err := Events(ctx, f.location, lo.ToPtr(numEventPages))
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

	// Keep length of feed to maximum number of items
	if len(f.feed.Items) > f.maxItems {
		f.feed.Items = f.feed.Items[:f.maxItems]
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

func NewFeed(location *string, maxSize *int) *Feed {
	feedTitle := "T4G Feed"
	feedDescription := "Tickets For Good Events"
	if location != nil {
		titleLocation := cases.Title(language.English).String(*location)

		feedTitle = fmt.Sprintf("%s: %s", feedTitle, titleLocation)
		feedDescription = fmt.Sprintf("%s in %s", feedDescription, titleLocation)
	}

	maxFeedItems := 100
	if maxSize != nil {
		maxFeedItems = *maxSize
	}

	return &Feed{
		location: location,
		maxItems: maxFeedItems,
		feed: &feeds.Feed{
			Title:       feedTitle,
			Link:        &feeds.Link{Href: EventsUrl(EventsInput{Location: location})},
			Description: feedDescription,
		},
	}
}

func feedSortFunc(item1, item2 *feeds.Item) bool {
	feedId1, err := strconv.Atoi(item1.Id)
	if err != nil {
		feedId1 = -1
	}

	feedId2, err := strconv.Atoi(item2.Id)
	if err != nil {
		feedId2 = -1
	}

	return feedId1 > feedId2
}
