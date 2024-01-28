FROM golang:1.21.5 AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

ENV GOARCH=amd64

RUN go build \
		-ldflags "-X main.buildCommit=`git rev-parse --short HEAD` \
		-X main.buildTime=`date "+%Y-%m-%dT%H:%M:%S%Z:00"`" \
		-o /go/bin/app

## deploy
FROM gcr.io/distroless/base-debian12

COPY --from=build /go/bin/app /app

EXPOSE 8081

USER nonroot:nonroot

CMD ["/app"]