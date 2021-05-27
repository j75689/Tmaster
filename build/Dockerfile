FROM golang:1.16 as builder

WORKDIR /Tmaster
COPY . /Tmaster

ENV CGO_ENABLED=0
ENV GO111MODULE=on

RUN go mod vendor
RUN go get github.com/google/wire/cmd/wire
RUN wire ./...
RUN go build -ldflags="-X github.com/j75689/Tmaster/cmd.version=$(git describe --tags) -X github.com/j75689/Tmaster/cmd.commitID=$(git rev-parse HEAD) -X github.com/j75689/Tmaster/cmd.commitDate=$(git log -n1 --pretty='format:%cd' --date=format:'%Y-%m-%d_%H:%M:%S')" -mod=readonly -v -o tmaster_bin main.go

FROM alpine:3

RUN apk add --no-cache ca-certificates

RUN mkdir -p /Tmaster
WORKDIR /Tmaster

COPY --from=builder /Tmaster/tmaster_bin /Tmaster/tmaster_bin

# Create appuser.
ENV USER=appuser
ENV UID=51807

RUN adduser \
--disabled-password \
--gecos "application user" \
--no-create-home \
--uid "${UID}" \
"${USER}"

RUN chown appuser:appuser /Tmaster
RUN chown appuser:appuser /Tmaster/*
USER appuser:appuser

ENTRYPOINT ["/Tmaster/tmaster_bin"]