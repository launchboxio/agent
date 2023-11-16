FROM golang:1.21

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./
COPY pkg/ pkg/
COPY cmd/ cmd/

ARG APP_VERSION="v0.0.0+unknown"
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/launchboxio/agent/cmd.version=$APP_VERSION" \
    -o /agent

ENTRYPOINT ["/agent"]
