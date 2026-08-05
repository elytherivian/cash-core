# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/cash ./cmd/cash

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/cash /cash
ENV DB_NAME=cash
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/cash"]
