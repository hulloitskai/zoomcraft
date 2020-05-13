#!/usr/bin/env sh

set -eo pipefail

docker-compose up -d minecraft

echo "!!! run 'cd client && yarn start' in another shell !!!" >&2
cat <<EOF | parallel --will-cite --ungroup "sh -c"
cd backend && go run .
cd gateway && yarn start
EOF
