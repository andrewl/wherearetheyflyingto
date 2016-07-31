#!/bin/bash
echo "Building"

if [ "${ARCH}" = "arm" ]; then
  env CGO_ENABLED=1 GOOS=linux GOARCH=arm go build
else
  go build
fi
