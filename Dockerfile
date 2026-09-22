FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app .

FROM alpine:3.21
RUN adduser -D app
USER app
COPY --from=build /app /app
EXPOSE 8080
CMD ["/app"]
