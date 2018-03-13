#!/bin/sh

nohup ./navi-agent -c cfg.json >agent.log 2>&1 &
nohup ./mytest 9292 >mytest.log 2>&1 &