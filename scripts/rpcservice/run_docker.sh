#!/bin/sh

     docker run -d --env-file=../env.ini -p9292:9292 --name=MyTest -v /opt/golang/src/kuaishangtong/navi/scripts/rpcservice/rpc:/rpc mytest:alpha
     docker run --env-file=../env.ini -p9292:9292 -v /opt/golang/src/kuaishangtong/navi/scripts/rpcservice/rpc:/rpc mytest:alpha
     #docker run -it --env-file=../env.ini -p9292:9292 -v /opt/golang/src/kuaishangtong/navi/scripts/rpcservice/rpc:/rpc mytest:alpha /bin/bash
     # /rpc/update_config.sh agent /rpc/etc/cfg.json
     # /rpc/run.sh