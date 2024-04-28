# syntax=docker/dockerfile:1
FROM golang:1.22 as builder

ENV GOOS=linux

COPY . .
# Build the Go app
RUN make build

ENTRYPOINT [ "./build/supernova" ]
