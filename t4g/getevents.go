package t4g

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ahobsonsayers/t4g-feed/utils"
)

const TicketsForGoodURL = "https://nhs.ticketsforgood.co.uk"

var ticketsForGoodUrl *url.URL

type EventsInput struct {
	Location *string
	Page     *int
}

func init() {
	var err error
	ticketsForGoodUrl, err = url.Parse(TicketsForGoodURL)
	if err != nil {
		log.Fatal("failed to parse tickets for good url")
	}
}

func EventsUrl(input EventsInput) string {
	ticketsForGoodEventsUrl := utils.CloneURL(ticketsForGoodUrl)
	ticketsForGoodEventsUrl.Path = "events"

	// Set query params
	queryParams := ticketsForGoodEventsUrl.Query()
	queryParams.Set("sort", "newest")
	queryParams.Set("range", "30")
	if input.Location != nil && *input.Location != "" {
		queryParams.Set("location", *input.Location)
	}
	if input.Page != nil && *input.Page > 0 {
		queryParams.Set("page", strconv.Itoa(*input.Page))
	}
	ticketsForGoodEventsUrl.RawQuery = queryParams.Encode()

	return ticketsForGoodEventsUrl.String()
}

func getEventsPage(ctx context.Context, input EventsInput) (string, error) {
	// Get page
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, EventsUrl(input), http.NoBody)
	if err != nil {
		return "", err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	httpError := utils.HTTPResponseError(response)
	if httpError != nil {
		return "", httpError
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}
