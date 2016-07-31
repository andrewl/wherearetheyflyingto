#!/bin/bash
IFS='-' read -ra PLATFORM <<< "$PLATFORM"
GOOS=${PLATFORM[0]}
GOARCH=${PLATFORM[1]}
COMPILER_ENV="CGO_ENABLED=1"
OUTPUT_FILENAME=wherearetheyflyingto-$GOOS-$GOARCH

echo "Building ${GOOS} ${GOARCH} to ${OUTPUT_FILENAME}"

if [ "${GOARCH}" = "arm" ]; then
  sudo apt-get install gcc-arm-linux-gnueabi 
  COMPILER_ENV="${COMPILER_ENV} CC=arm-linux-gnueabi-gcc"
fi

env ${COMPILER_ENV} GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${OUTPUT_FILENAME}
