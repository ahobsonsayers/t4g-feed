package t4g

import (
	"context"
	"strings"
	"time"

	"github.com/ahobsonsayers/t4g-feed/utils"
	"github.com/foolin/pagser"
	"github.com/gorilla/feeds"
)

type T4G struct {
	Events []Event `pagser:"[class*='event_card']"`
}

type Event struct {
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

func (e Event) ToFeedItem() *feeds.Item {
	return &feeds.Item{
		Title:       e.Title,
		Link:        &feeds.Link{Href: e.Link},
		Description: strings.Join([]string{e.Date, e.Location, e.Category}, "\n"),
		Enclosure:   &feeds.Enclosure{Url: e.Image, Type: "image/jpeg", Length: "0"},
		Created:     time.Now(),
		Updated:     time.Now(),
	}
}

func Events(ctx context.Context, location *string) ([]Event, error) {
	// Get events page
	eventsPage, err := getEventsPage(ctx, location)
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
