###
# Builder to compile our golang code
###
FROM golang:alpine AS builder

WORKDIR /build
COPY . .

RUN go install -v

###
# Now generate our smaller image
###
FROM alpine

COPY --from=builder /go/bin/absol /go/bin/absol

CMD ["/go/bin/absol"]
