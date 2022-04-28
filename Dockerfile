###
# Builder to compile our golang code
###
FROM golang:1.18-alpine AS builder

WORKDIR /build
COPY . .

RUN go build -o absol -buildvcs=false -v github.com/lordralex/absol/core

###
# Now generate our smaller image
###
FROM alpine

COPY --from=builder /build/absol /go/bin/absol

ENV DISCORD_TOKEN="YOUR DISCORD BOT TOKEN"
ENV DATABASE=""

ENTRYPOINT ["/go/bin/absol"]
CMD ["alert", "cleaner", "factoids", "log", "twitch", "hjt", "mcping"]
