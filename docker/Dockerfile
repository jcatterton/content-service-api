FROM golang:1.15.8-alpine as builder
RUN apk update && apk upgrade && apk add --no-cache bash libc6-compat git openssh
WORKDIR /content-service-api
COPY . .
RUN rm -f go.sum
RUN go build -o ./app ./cmd/svr/main.go

FROM alpine:3.13.1
RUN apk update && apk upgrade && apk add ffmpeg
WORKDIR /app
COPY --from=builder /content-service-api/app .
EXPOSE 8005
CMD ["./app"]
