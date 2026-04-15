#!/usr/bin/env bash
set -e
cd "$(dirname "$0")"

# Build the binary
go build -o ~/.local/bin/jjui-ws ./cmd/jjui

# The nix devshell in this directory puts an older jj on PATH.
# Strip nix store paths from PATH so jjui picks up the user's jj.
CLEAN_PATH=$(echo "$PATH" | tr ':' '\n' | grep -v '/nix/store/' | tr '\n' ':' | sed 's/:$//')

exec env PATH="$CLEAN_PATH" ~/.local/bin/jjui-ws "$@"
