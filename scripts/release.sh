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
  -output="pkg/{{.Dir}}-${version}-{{.OS}}-{{.Arch}}/gotee"

pushd pkg
  for dir in $(ls -1 .); do
    shafile="${dir}.sha"
    zipfile="${dir}.zip"
    zip -j "${zipfile}" "${dir}/gotee"
    shasum -a 256 "${zipfile}" > "${shafile}"
  done
popd

ghr -recreate ${version} pkg
