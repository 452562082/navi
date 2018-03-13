#!/bin/sh

nohup ./navi-agent -c cfg.json >agent.log 2>&1 &
nohup ./mytest >mytest.log 2>&1 &