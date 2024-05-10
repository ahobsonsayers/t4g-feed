# t4g-feed

A project to provide an RSS feed for [Tickets For Good](https://ticketsforgood.co.uk/) 🎉

This allows you to use any standard RSS application to see listed events on Tickets For Good, and more importantly, get **notified** when new events are posted 🤩

A great (and recommended) Android app to use this feed is [Feeder](https://play.google.com/store/apps/details?id=com.nononsenseapps.feeder.play)

## Usage

You can use the feed (e.g. for London) by entering the following URL into your RSS app:

https://ticketsforgood.co.uk/london

You can also substitute **london** for any location in the UK, or even a postcode. It works like the normal location search on the Tickets For Good website:

https://ticketsforgood.co.uk/<location>

Notes:

- The default (and currently **only**) search radius around the chosen location is 30 miles.
- The feed will update every 5 minutes upon request to the server.

## Run it yourself

You can run this RSS feed server yourself using a handy Docker container provided. You can run it with the following command:

```bash
docker run -d \
  --name t4g-feed \
  -p 5656:5656 \
  --restart unless-stopped \
  arranhs/t4g-feed:develop
```
