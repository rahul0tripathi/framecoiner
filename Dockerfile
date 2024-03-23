FROM golang:1.22-alpine AS builder

# Install git + SSL ca certificates.
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache make git tzdata ca-certificates && update-ca-certificates
# RUN apk update && apk add --no-cache make git tzdata
ENV USER=appuser
ENV UID=10001

WORKDIR $GOPATH/src/

# See https://stackoverflow.com/a/55757473/12429735
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

COPY . .

RUN make build

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

# Copy our static executable
COPY --from=builder /go/src/_bin/framecoinder /go/bin/framecoinder

# Use an unprivileged user.
USER appuser:appuser

EXPOSE 8000
ENTRYPOINT ["/go/bin/framecoinder"]