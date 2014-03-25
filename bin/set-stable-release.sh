#!/bin/bash

set -e

if [ -z "$AWS_ACCESS_KEY_ID" ]; then
    echo "Need to set AWS_ACCESS_KEY_ID"
    exit 1
fi

if [ -z "$AWS_SECRET_ACCESS_KEY" ]; then
    echo "Need to set AWS_SECRET_ACCESS_KEY"
    exit 1
fi

VERSION=$1
if [ -z "$VERSION" ]; then
    echo "Usage: set-stable-release VERSION"
    echo "Example: set-stable-release v6.1.1"
    exit 1
fi

root_dir=$(cd $(dirname $0) && pwd)/..
s3_config_file=$root_dir/ci/s3cfg

mkdir -p tmp
cd tmp
touch empty-file

files=(
 cf-linux-amd64.tgz
 cf-linux-386.tgz
 cf-darwin-amd64.tgz
 cf-windows-amd64.zip
 cf-windows-386.zip
 cf-cli_amd64.deb
 cf-cli_i386.deb
 cf-cli_amd64.rpm
 cf-cli_i386.rpm
 installer-osx-amd64.pkg
 installer-windows-amd64.zip
 installer-windows-386.zip
)

for file in ${files[*]}
do
  echo "uploading file" $file
  header="x-amz-website-redirect-location:/releases/$VERSION/$file"
  s3cmd put empty-file s3://go-cli/releases/latest/$file --config=$s3_config_file --add-header $header > /dev/null 2>&1
done
