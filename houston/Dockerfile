# syntax=docker/dockerfile:1
FROM golang:1.22 as builder
ENV GOOS=linux

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o ./bin/houston main.go

FROM golang:1.22 as run-houston

# CREATE USER & CHANGE DIR
ENV USR=go_user
ENV GRP=go_grp

RUN set -x; addgroup --gid 1000 --system "$GRP" \
  && adduser --disabled-password --system --home "/home/$USR" \
  --uid 1000 --gid 1000 "$USR"

# reduce privileges
USER $USR
WORKDIR /go/bin

COPY --from=builder --chown=1000:1000 /go/bin/houston .
ENTRYPOINT [ "./houston" ]
