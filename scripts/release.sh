#!/bin/bash

rm -rf release
mkdir release

go install

v="$(taho -v)"

for os in linux darwin; do
  for arch in amd64 arm64; do
    env GOOS="$os" GOARCH="$arch" go build .
    mv taho "release/taho-$v-$os-$arch"
  done
done
