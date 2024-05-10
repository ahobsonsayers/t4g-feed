# t4g-feed

## Running

```bash
docker run -d \
    --name t4g-feed \
    -p 5656:5656 \
    --restart unless-stopped \
    arranhs/t4g-feed:develop
```
