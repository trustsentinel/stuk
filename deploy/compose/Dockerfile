# Build stuk binaries, then a small runtime with sshd + iptables for the demo.
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN set -eux; for c in stukd stuk stuk-authkeys; do \
      CGO_ENABLED=0 go build -o "/out/$c" "./cmd/$c"; \
    done

FROM alpine:3.24
RUN apk add --no-cache openssh openssh-client iptables conntrack-tools \
 && mkdir -p /var/empty /run/sshd /keys /run/stuk/keys
COPY --from=build /out/stukd /out/stuk /out/stuk-authkeys /usr/local/bin/
COPY deploy/compose/gateway-entrypoint.sh deploy/compose/client-entrypoint.sh \
     deploy/compose/stukd.json /app/
RUN chmod +x /app/*.sh
WORKDIR /app
