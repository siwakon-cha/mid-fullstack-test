FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/backend .

FROM build AS test
CMD ["go", "test", "-tags=integration", "./...", "-count=1", "-v"]

FROM alpine:3.22 AS runtime
RUN apk add --no-cache ca-certificates wget && addgroup -S app && adduser -S -G app app
COPY --from=build /out/backend /usr/local/bin/backend
USER app
EXPOSE 8080
ENTRYPOINT ["backend"]
