#!/bin/sh

cat <<EOF | parallel --will-cite --ungroup
node gateway/server.js
./backend/backend
EOF
fi
