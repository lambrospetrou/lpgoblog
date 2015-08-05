#!/bin/bash

BUILD_DIR="build-release"

# update sources
#go -u get
go get
# build the executable
go build -o goblog

# create the release folder and put the required files inside
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

cp -r goblog deploy_linode.sh static/ templates/ sec/ "$BUILD_DIR/"

# zip the bundle into a single file
tar -cvzf goblog.tgz "$BUILD_DIR"
mv goblog.tgz "$BUILD_DIR/"

