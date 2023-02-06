#!/bin/sh

# Registry
kubectl port-forward --namespace kube-system --address 0.0.0.0 service/registry 5000:80 &

# Nats
kubectl port-forward --address 0.0.0.0 service/ipfs-telemetry-nats 4222 &

# TimescaleDB
kubectl port-forward --address 0.0.0.0 service/ipfs-telemetry-tsdb 5432 &

# Redis
kubectl port-forward --address 0.0.0.0 service/ipfs-telemetry-redis-master 6379 &

# VictoriaMetrics
kubectl port-forward --address 0.0.0.0 service/ipfs-telemetry-vm-server 8428 &

echo "Port forwarding..."
echo "Press Ctrl+C to stop"
wait