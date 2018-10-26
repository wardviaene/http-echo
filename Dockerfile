#
# Build go project
#
FROM golang:1.11-alpine as go-builder

WORKDIR /go/src/github.com/in4it/http-echo/

COPY . .

RUN apk add -u -t build-tools curl git && \
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && \
    dep ensure && \
    apk del build-tools && \
    rm -rf /var/cache/apk/*

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o http-echo *.go

#
# Runtime container
#
FROM alpine:latest  

RUN mkdir -p /app && \
    addgroup -S app && adduser -S app -G app && \
    chown app:app /app

WORKDIR /app

COPY --from=go-builder /go/src/github.com/in4it/http-echo/http-echo .

USER app

CMD ["./http-echo"]  
