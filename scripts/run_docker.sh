#!/bin/sh

docker run -d --env-file=env.ini -p9292:9292 --name:MyTest mytest:alpha
#docker run -it --env-file=env.ini -p9292:9292 mytest:alpha /bin/bash
# /rpc/update_config.sh agent /rpc/etc/cfg.json
# /rpc/run.sh