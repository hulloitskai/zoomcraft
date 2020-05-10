#!/bin/sh

set -euo pipefail

mock() {
  local PACKAGE="$(go list .)"
  local DESTINATION=${2:-mock_gen.go}

  go run github.com/golang/mock/mockgen \
    -destination $DESTINATION \
    -package "$(basename $PACKAGE)" \
    -self_package "$PACKAGE" \
    . $1
}

mock $@
