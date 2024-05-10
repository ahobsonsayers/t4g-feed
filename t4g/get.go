package t4g

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/ahobsonsayers/t4g-feed/utils"
)

const TicketsForGoodURL = "https://nhs.ticketsforgood.co.uk"

var (
	ticketsForGoodUrl       *url.URL
	ticketsForGoodEventsUrl *url.URL
)

func init() {
	var err error

	ticketsForGoodUrl, err = url.Parse(TicketsForGoodURL)
	if err != nil {
		log.Fatal("failed to parse tickets for good url")
	}

	ticketsForGoodEventsUrl = utils.CloneURL(ticketsForGoodUrl)
	ticketsForGoodEventsUrl.Path = "events"
	queryParams := ticketsForGoodEventsUrl.Query()
	queryParams.Set("sort", "newest")
	ticketsForGoodEventsUrl.RawQuery = queryParams.Encode()
}

func LocationEventsUrl(location *string) string {
	if location == nil {
		return ticketsForGoodEventsUrl.String()
	}

	ticketsForGoodLocationEventsUrl := utils.CloneURL(ticketsForGoodEventsUrl)
	queryParams := ticketsForGoodLocationEventsUrl.Query()
	queryParams.Set("location", *location)
	ticketsForGoodLocationEventsUrl.RawQuery = queryParams.Encode()

	return ticketsForGoodLocationEventsUrl.String()
}

func getEventsPage(ctx context.Context, location *string) (string, error) {
	// Get page
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, LocationEventsUrl(location), http.NoBody)
	if err != nil {
		return "", err
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", nil
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
