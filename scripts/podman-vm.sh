#!/usr/bin/env -S bash -x
set -e
cd $(dirname $0)/..

# uses port 8428

mkdir -p data/vm
podman run -d --name vm --network host \
    -v ./data/vm:/victoria-metrics-data:z \
    -v ./scripts/scrape.yaml:/etc/scrape.yaml:ro,z \
    docker.io/victoriametrics/victoria-metrics:latest \
    --selfScrapeInterval=5s -storageDataPath=/victoria-metrics-data \
    -promscrape.config=/etc/scrape.yaml
