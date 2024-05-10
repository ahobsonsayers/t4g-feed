package server

import (
	"context"
	"strings"
	"time"

	"github.com/ahobsonsayers/t4g-feed/t4g"
)

type server struct{}

func NewServer() StrictServerInterface { return &server{} }

func (*server) T4g(ctx context.Context, request T4gRequestObject) (T4gResponseObject, error) {
	feed, err := t4g.FeedDebounce(ctx, &request.Location, time.Minute)
	if err != nil {
		return T4g400JSONResponse{ErrorJSONResponse{Error: err.Error()}}, nil
	}

	rssFeed, err := feed.ToRss()
	if err != nil {
		return T4g400JSONResponse{ErrorJSONResponse{Error: err.Error()}}, nil
	}

	rssFeedReader := strings.NewReader(rssFeed)

	return T4g200ApplicationxmlResponse{
		Body:          rssFeedReader,
		ContentLength: int64(rssFeedReader.Len()),
	}, nil
}
