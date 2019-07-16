FROM golang:alpine

RUN mkdir -p /go/src/github.com/lordralex/absol
WORKDIR /go/src/github.com/lordralex/absol

COPY . .

RUN apk add --no-cache curl git iputils bash \
    && curl -L -o /go/bin/dep https://github.com/golang/dep/releases/download/v0.5.1/dep-linux-amd64 \
    && chmod +x /go/bin/dep \
    && dep ensure -v \
    && go install -v github.com/lordralex/absol \
    && apk del curl git

CMD ["absol"]