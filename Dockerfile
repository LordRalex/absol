###
# Builder to compile our golang code
###
FROM golang:1.24-alpine AS builder

ARG tags="modules.all,databases.all"

WORKDIR /build
COPY . .

RUN go build -o absol -buildvcs=false -tags=$tags -v github.com/lordralex/absol/core

###
# Now generate our smaller image
###
FROM alpine

COPY --from=builder /build/absol /go/bin/absol

ENV DISCORD_TOKEN="YOUR DISCORD BOT TOKEN"
ENV DATABASE_DIALECT="mysql"
ENV DATABASE_URL=""

ENTRYPOINT ["/go/bin/absol"]
CMD ["all"]
