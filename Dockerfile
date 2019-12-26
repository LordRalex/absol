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
FROM python:3-alpine

RUN pip install mcstatus
COPY --from=builder /go/bin/absol /go/bin/absol
COPY --from=builder /build/handlers/mcping/mcping.py /go/bin/mcping.py

CMD ["/go/bin/absol"]
