#!/bin/sh

/rpc/update_config.sh navi /rpc/etc/service.yaml

echo "start navi-httpservice"

/rpc/bin/MyTest -path ./service.yaml
