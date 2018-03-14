#!/bin/sh

/navi/update_config.sh navi /navi/etc/service.yaml
echo "start navi-httpservice"
/navi/bin/MyTest -path /navi/etc/service.yaml
