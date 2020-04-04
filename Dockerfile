FROM golang:alpine AS build_stage
RUN apk update && apk add git && rm -rf /var/cache/apk/*
COPY ./src ./src
RUN go get -d -v ./src/ && go build -o ./bin/serviceNowFlag -v ./src/

FROM alpine:latest
WORKDIR /root/
COPY --from=build_stage /go/bin/serviceNowFlag .
ENTRYPOINT ["./serviceNowFlag"]