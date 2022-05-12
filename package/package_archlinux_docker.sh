#!/usr/bin/env bash

docker run --rm -e PACKAGE_ARCH=amd64 -v $(pwd):/project:Z archlinux:latest bash -c '/project/package/package_archlinux_docker_entry.sh'
