#!/bin/bash

set -e

version=$1

if [[ -z "${version}" ]]; then
  >&2 echo "usage: scripts/release.sh v0.0.0"
  exit 1
fi

XC_ARCH=${XC_ARCH:-386 amd64}
XC_OS=${XC_OS:-darwin linux}

rm -rf pkg/
gox \
  -os="${XC_OS}" \
  -arch="${XC_ARCH}" \
  -output="pkg/{{.Dir}}_${version}_{{.OS}}_{{.Arch}}/gotee"

ls -1 pkg | xargs -t -I{} zip -j pkg/{}.zip pkg/{}/gotee

shasum -a 256 pkg/*.zip | sed 's: pkg/::' > pkg/${version}_sums.txt
ghr -recreate -replace ${version} pkg
