FROM golang:1.25 AS build
WORKDIR /src
COPY . .
RUN go build -o /bin/app ./...

FROM gcr.io/distroless/base
COPY --from=build /bin/app /app
ENTRYPOINT ["/app"]
