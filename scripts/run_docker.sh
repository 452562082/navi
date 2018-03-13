#!/bin/sh

docker run -d --env-file=env.ini -p9292:9292 mytest:alpha
#docker run -it --env-file=env.ini -p9292:9292 mytest:alpha /bin/sh
# /rpc/update_config.sh agent /rpc/etc/cfg.json
# /rpc/run.sh