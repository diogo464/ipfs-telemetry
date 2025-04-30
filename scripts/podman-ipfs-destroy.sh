#!/bin/bash

echo "Stopping and removing all kubo-* containers..."
podman ps -a --format "{{.Names}}" | grep '^kubo-' | xargs -r podman stop
echo "All kubo containers stopped."
