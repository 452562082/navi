#!/bin/sh

docker run --env-file=../env.ini -p8081:8081 -v /opt/golang/src/kuaishangtong/navi/scripts/httpservice/navi:/navi mytest_http:alpha

