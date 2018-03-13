#!/bin/sh

export LD_LIBRARY_PATH=/usr/local/lib:/rpc/libs

/rpc/update_config.sh agent /rpc/etc/cfg.json

echo "start mytest"

nohup /rpc/bin/mytest 9292 >/rpc/logs/mytest.log 2>&1 &
sleep 3
echo "start navi-agent"

/rpc/bin/navi-agent -c /rpc/etc/cfg.json