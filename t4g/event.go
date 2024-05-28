package t4g

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/ahobsonsayers/t4g-feed/utils"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/samber/lo"
)

const EventsPerPage = 12

type T4G struct {
	Events []Event `pagser:"[class*='event_card']"`
}

type Event struct {
	Id       int    `pagser:".card-body a->attrNumbers(href)"` // Id can be found in the link
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

	event.Image = strings.ReplaceAll(e.Image, "thumb_", "")

	link := utils.CloneURL(ticketsForGoodUrl)
	link.Path = e.Link
	event.Link = link.String()

	event.Date = strings.ReplaceAll(e.Date, "\n", " ")

	return event
}

func Events(ctx context.Context, location *string, pages *int) ([]Event, error) {
	eventPages, err := manyPageEvents(ctx, location, pages)
	if err != nil {
		return nil, err
	}

	// Pages have a maximum of 12 items
	maxEventCount := 12 * len(eventPages)

	// Flatten and dedupe events
	events := make([]Event, 0, maxEventCount)
	seenEventIds := mapset.NewSetWithSize[int](maxEventCount)
	for _, pageEvents := range eventPages {
		for _, event := range pageEvents {
			if event.Id != 0 && seenEventIds.Contains(event.Id) {
				continue
			}

			events = append(events, event)
			seenEventIds.Add(event.Id)
		}
	}

	return events, nil
}

func pageEvents(ctx context.Context, input EventsInput) ([]Event, error) {
	// Get events page
	eventsPage, err := getEventsPage(ctx, input)
	if err != nil {
		return nil, err
	}

	// Parse t4g
	var t4g T4G
	err = NewHTMLParser().Parse(&t4g, eventsPage)
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

func manyPageEvents(ctx context.Context, location *string, numPages *int) ([][]Event, error) {
	pages := lo.FromPtr(numPages)
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
