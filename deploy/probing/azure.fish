#!/bin/fish

set IMAGE "ghcr.io/diogo464/probing:latest"
set PREFIX "probe"
set RESOURCE_GROUP "containers"
set LOCATIONS "japaneast" "eastus" "southindia" "switzerlandwest"

if [ "$argv[1]" = "up" ]
    for LOC in $LOCATIONS
        az container create --resource-group containers --name "$PREFIX-$LOC" --image "$IMAGE" --cpu 1 --memory 1 --ports 4640 --command-line "/bin/sh -c './probe --name $LOC'" --ip-address Public --location $LOC
    end
else if [ "$argv[1]" = "down" ]
    for LOC in $LOCATIONS
        az container delete --yes --resource-group $RESOURCE_GROUP --name "$PREFIX-$LOC"
    end
else if [ "$argv[1]" = "ip" ]
    az container list | jq '.[] | select(.containers[].name | contains("probe")) | .ipAddress.ip' | jq -r
else
    echo "invalid usage"
end
