#!/bin/bash

podman ps --format "{{.Names}}" | grep '^kubo-' | while read -r name; do
  n="${name#kubo-}"
  echo "127.0.0.1:$((30000 + n))"
done
