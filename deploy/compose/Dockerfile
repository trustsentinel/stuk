# Build stuk binaries, then a small runtime with sshd + iptables for the demo.
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/stukd ./cmd/stukd \
 && CGO_ENABLED=0 go build -o /out/stuk  ./cmd/stuk

FROM alpine:3.20
RUN apk add --no-cache openssh openssh-client iptables conntrack-tools \
 && mkdir -p /var/empty /run/sshd /keys
COPY --from=build /out/stukd /out/stuk /usr/local/bin/
COPY deploy/compose/gateway-entrypoint.sh deploy/compose/client-entrypoint.sh \
     deploy/compose/grant.sh deploy/compose/revoke.sh deploy/compose/stukd.json /app/
RUN chmod +x /app/*.sh
WORKDIR /app
