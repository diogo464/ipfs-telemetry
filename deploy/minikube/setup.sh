#!/bin/sh

cd $(dirname $0)

echo "Installing Prometheus Operator"
LATEST=$(curl -s https://api.github.com/repos/prometheus-operator/prometheus-operator/releases/latest | jq -cr .tag_name)
curl -sL https://github.com/prometheus-operator/prometheus-operator/releases/download/${LATEST}/bundle.yaml | kubectl create -f -

echo "Installing Grafana Operator"
helm install -f values/grafana-operator.yaml grafana-operator bitnami/grafana-operator

echo "Applying resources"
kubectl apply -f resources/

echo "Installing ipfs-telemetry"
helm install -f values/ipfs-telemetry.yaml ipfs-telemetry ../chart/