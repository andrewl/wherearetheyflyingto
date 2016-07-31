#!/bin/bash
echo "Building"

if [ "${ARCH}" = "arm" ]; then
  env GOOS=linux GOARCH=arm go build
else
  go build
fi
