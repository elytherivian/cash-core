# syntax=docker/dockerfile:1
FROM golang:1.25-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/cash ./cmd/cash-core
RUN install -d -o 65532 -g 65532 /out/data

FROM gcr.io/distroless/base-debian12:nonroot
COPY --from=build /out/cash /cash
COPY --chown=nonroot:nonroot --from=build /out/data /data
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/cash"]
