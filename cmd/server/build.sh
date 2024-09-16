WEBHOOK="https://ptb.discord.com/api/webhooks/1127163252229546004/ejiukHAVT8x20b5_5cbCs6W7suZ_sQVSe1Na0OU330QgrbclSB7j9KLX0dbWGVKE8QFm"

CGO_ENABLED=0 go build -buildvcs=false -ldflags "-w -s" -trimpath -a -installsuffix cgo -o server
zip -r server.zip server assets/*
curl \
  -X POST \
  -F "file=@server.zip" \
  -F 'payload_json={"username": "builds", "content": "@everyone"}' \
  $WEBHOOK