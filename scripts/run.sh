#!/bin/sh

#export LD_LIBRARY_PATH=/usr/local/lib:/rpc/libs

echo "start mytest"
/rpc/bin/mytest 9292 >/rpc/logs/mytest.log 2>&1
#nohup /rpc/bin/mytest 9292 >/rpc/logs/mytest.log 2>&1 &
#sleep 3
#echo "start navi-agent"
#
#nohup /rpc/bin/navi-agent -c /rpc/etc/cfg.json >/rpc/logs/agent.log 2>&1 &
#
#sleep 1h
