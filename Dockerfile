FROM golang:alpine as builder
COPY .  /webapp
RUN apk update && apk add git &&\
    go get -d -v  github.com/boltdb/bolt &&\
    go get -d -v  github.com/go-chi/chi &&\
    go get -d -v  github.com/json-iterator/go &&\
    cd /webapp && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /webapp/apk
    	
FROM alpine:latest
EXPOSE 8080
VOLUME ["/webapp"]
COPY --from=builder /webapp/apk /webapp/apk
ENTRYPOINT [ "/webapp/apk" ]
