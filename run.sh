#!/usr/bin/env bash
set -e
cd "$(dirname "$0")"
go build -o ~/.local/bin/jjui-ws ./cmd/jjui
exec ~/.local/bin/jjui-ws "$@"
