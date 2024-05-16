package t4g_test

import (
	"context"
	"net/url"
	"testing"

	"github.com/ahobsonsayers/t4g-feed/t4g"
	"github.com/stretchr/testify/require"
)

func TestT4GEvents(t *testing.T) {
	events, err := t4g.Events(context.Background(), nil, nil)
	require.NoError(t, err)
	require.Len(t, events, 12)

	require.NotEmpty(t, events[0].Id)
	require.NotEmpty(t, events[0].Title)
	require.NotEmpty(t, events[0].Image)
	require.NotEmpty(t, events[0].Link)
	require.NotEmpty(t, events[0].Location)
	require.NotEmpty(t, events[0].Date)
	require.NotEmpty(t, events[0].Category)

	_, err = url.Parse(events[0].Image)
	require.NoError(t, err)

	_, err = url.Parse(events[0].Link)
	require.NoError(t, err)
}
