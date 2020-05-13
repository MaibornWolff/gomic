FROM golang:alpine as builder
RUN apk update
RUN apk add gcc g++
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN go build -o main .
FROM alpine
RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /build/ /app/
WORKDIR /app
CMD ["./main"]
