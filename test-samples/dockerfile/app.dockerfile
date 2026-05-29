FROM alpine:3.20

WORKDIR /app

COPY . .

RUN echo "building" && ls -la

ENV APP_ENV=production
EXPOSE 8080

CMD ["sh", "-c", "echo hello"]
