#!/bin/bash

COUNT=$1  # Number of manifests to generate
ADDRESS="10.0.0.160"

if [ -z "$COUNT" ]; then
  echo "Usage: $0 <number_of_manifests>"
  exit 1
fi

for ((n=0; n<COUNT; n++)); do
  PORT=$((29000 + n))
  NAME_TCP="kubo-tcp-$n"
  NAME_UDP="kubo-udp-$n"

  # Print Kubernetes manifest
  echo "---"
  cat <<EOF
apiVersion: infra.d464.sh/v1
kind: PortForward
metadata:
  namespace: dev
  name: '$NAME_TCP'
spec:
  address: $ADDRESS
  port: $PORT
  protocol: TCP
EOF
  echo "---"
  cat <<EOF
apiVersion: infra.d464.sh/v1
kind: PortForward
metadata:
  namespace: dev
  name: '$NAME_UDPTCP'
spec:
  address: $ADDRESS
  port: $PORT
  protocol: UDP
EOF
done
