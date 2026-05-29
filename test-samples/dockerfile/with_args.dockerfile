ARG GO_VERSION=1.25

FROM golang:${GO_VERSION} AS build
WORKDIR /src
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/app ./...

FROM alpine:3.20
ENV APP_ENV=production
RUN apk add --no-cache ca-certificates
COPY --from=build /bin/app /usr/local/bin/app
HEALTHCHECK --interval=30s CMD ["app", "--health"]
USER 1000:1000
EXPOSE 8080
ENTRYPOINT ["app"]
