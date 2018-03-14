#!/bin/sh

docker run --env-file=../env.ini -p9292:9292 --name=MyTest -v /opt/golang/src/kuaishangtong/navi/scripts/rpc:/rpc mytest_http:alpha
