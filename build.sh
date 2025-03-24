#!/bin/dash

go clean -modcache

TAGS="with_gvisor,with_quic,with_wireguard,with_ech,with_utls,with_clash_api"
LDFLAGS="-s -w -buildid="

go build -v \
    -trimpath \
    -buildvcs=false \
    -tags="$TAGS" \
    -ldflags="$LDFLAGS" \
    -o elon \
    ./cmd/elon