#!/bin/bash
echo "Building"

if [ "${ARCH}" = "arm" ]; then
  sudo apt-get install gcc-arm-linux-gnueabi 
  env CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm go build -o wherearetheyflyingto-arm
else
  go build
fi
