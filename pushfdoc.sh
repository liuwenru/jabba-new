#!/bin/bash

make build-release
cp -rf release/jabba-0.14.0-linux-amd64 /root/remotefs/fdoc.liuwenru/tools/jabba/
cp -rf release/jabba-0.14.0-linux-arm64 /root/remotefs/fdoc.liuwenru/tools/jabba/
cp -rf local-install.sh /root/remotefs/fdoc.liuwenru/tools/jabba/
