FROM golang:alpine AS build_stage
RUN apk update && apk add git && rm -rf /var/cache/apk/*
WORKDIR /go/src
COPY ./src .
RUN go mod init main && go mod tidy && go get -d -v . && go build -o /go/bin/serviceNowFlag -v .

FROM alpine:latest
WORKDIR /root/
COPY --from=build_stage /go/bin/serviceNowFlag .
ENTRYPOINT ["./serviceNowFlag"]
