FROM golang:alpine

COPY . .

RUN apk add --no-cache curl git iputils bash \
    && go install -v github.com/lordralex/absol \
    && apk del curl git

CMD ["absol"]