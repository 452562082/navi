#!/bin/sh

nohup /go/src/mytest/navi-agent -c /go/src/mytest/cfg.json >/go/src/mytest/logs/agent.log 2>&1 &
nohup /go/src/mytest/mytest 9292 >/go/src/mytest/logs/mytest.log 2>&1 &