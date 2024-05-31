FROM golang:1.22 as build
WORKDIR /app/src

ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/root/.cache/go-build
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

# Leverage build cache by copying go.mod and go.sum first
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download
RUN go mod verify

# Copy the rest of the source code
COPY . .

RUN --mount=type=cache,target="/root/.cache/go-build" go build -o ../client ./cmd/client

FROM alpine as app

# Create a non-root user
ARG user=app
ARG group=docker
RUN addgroup $group
RUN adduser \
    --disabled-password \
    --gecos "" \
    --ingroup $group \
    $user
USER $user
WORKDIR /usr/local/bin/client

# Copy the binary and set executable permissions
COPY --chown=$user:$group --from=build /app/client /usr/local/bin/cli
RUN chmod a+x /usr/local/bin/cli

CMD ["/usr/local/bin/cli"]