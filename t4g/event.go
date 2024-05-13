package t4g

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ahobsonsayers/t4g-feed/utils"
	"github.com/foolin/pagser"
	"github.com/gorilla/feeds"
	"github.com/samber/lo"
)

// eventIdRegex is a regex to match an event id in a
// event link. Id is the first number in the link.
var eventIdRegex = regexp.MustCompile(`\d+`)

type T4G struct {
	Events []Event `pagser:"[class*='event_card']"`
}

type Event struct {
	Id       string `pagser:".card-body a->attr(href)"` // Id can be found in the link
	Title    string `pagser:".card-title"`
	Image    string `pagser:"img->attr(src)"`
	Link     string `pagser:".card-body a->attr(href)"`
	Location string `pagser:".card-body .col->eq(0)"`
	Date     string `pagser:".card-body .col->eq(1)"`
	Category string `pagser:".card-body .col->eq(2)"`
}

// sanitise will sanitise an event after being parsed from html
func (e Event) sanitise() Event {
	event := e

	event.Id = eventIdRegex.FindString(e.Id)

	event.Image = strings.ReplaceAll(e.Image, "thumb_", "")

	link := utils.CloneURL(ticketsForGoodUrl)
	link.Path = e.Link
	event.Link = link.String()

	event.Date = strings.ReplaceAll(e.Date, "\n", " ")

	return event
}

func (e Event) ToFeedItem() *feeds.Item {
	return &feeds.Item{
		Id:          e.Id,
		Title:       e.Title,
		Link:        &feeds.Link{Href: e.Link},
		Description: fmt.Sprintf("%s | %s | %s", e.Date, e.Location, e.Category),
		Enclosure:   &feeds.Enclosure{Url: e.Image, Type: "image/jpeg", Length: "0"},
		Created:     time.Now(),
	}
}

func Events(ctx context.Context, location *string) ([]Event, error) {
	eventPages, err := manyPageEvents(ctx, location, 5)
	if err != nil {
		return nil, err
	}
	return lo.Flatten(eventPages), nil
}

func pageEvents(ctx context.Context, input EventsInput) ([]Event, error) {
	// Get events page
	eventsPage, err := getEventsPage(ctx, input)
	if err != nil {
		return nil, err
	}

	// Parse t4g
	var t4g T4G
	err = pagser.New().Parse(&t4g, eventsPage)
	if err != nil {
		return nil, err
	}

	// Sanitise events
	events := make([]Event, 0, len(t4g.Events))
	for _, event := range t4g.Events {
		events = append(events, event.sanitise())
	}

	return events, nil
}

func manyPageEvents(ctx context.Context, location *string, pages int) ([][]Event, error) {
	if pages < 1 {
		pages = 1
	}

	eventPages := make([][]Event, pages)
	var errs error

	var wg sync.WaitGroup
	var eventsMutex sync.Mutex
	var errsMutex sync.Mutex

	for idx := 0; idx < pages; idx++ {
		wg.Add(1)

		go func(idx int) {
			defer wg.Done()

			pageEvents, err := pageEvents(
				ctx, EventsInput{
					Location: location,
					Page:     lo.ToPtr(idx + 1),
				},
			)
			if err != nil {
				errsMutex.Lock()
				errs = errors.Join(errs, err)
				errsMutex.Unlock()
				return
			}

			eventsMutex.Lock()
			eventPages[idx] = pageEvents
			eventsMutex.Unlock()
		}(idx)
	}
	wg.Wait()

	return eventPages, errs
}
