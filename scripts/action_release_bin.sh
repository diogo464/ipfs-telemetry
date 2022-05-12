#!/usr/bin/env bash

UPLOAD_URL=$(echo $UPLOAD_URL | cut -d'{' -f1)
for FILE in $(ls -1 bin/); do
    curl \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Content-Type: application/octet-stream" \
        --data-binary @bin/$FILE \
        "$UPLOAD_URL?name=$FILE"
done
