#!/usr/bin/env -S bash -x
set -e

COUNT=$1  # Number of containers to start

if [ -z "$COUNT" ]; then
  echo "Usage: $0 <number_of_containers>"
  exit 1
fi
podman build -t kubo -f Containerfile .
for ((n=0; n<COUNT; n++)); do
  PORT=$((29000 + n))
  API_PORT=$((30000 + n))
  GATEWAY_PORT=$((31000 + n))
  NAME="kubo-$n"

  echo "Starting container: $NAME (PORT=$PORT, API_PORT=$API_PORT, GATEWAY_PORT=$GATEWAY_PORT)"
  podman run --rm -itd --name "$NAME" \
    -e PORT="$PORT" \
    -e API_PORT="$API_PORT" \
    -e GATEWAY_PORT="$GATEWAY_PORT" \
    --network host \
    kubo
done

echo "All containers started."
