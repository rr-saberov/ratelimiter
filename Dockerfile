FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS build

ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/ratelimiter ./cmd/ratelimiter

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/ratelimiter /ratelimiter

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/ratelimiter"]
