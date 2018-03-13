#!/bin/sh

docker run -d --env-file=env.ini -p9292:9292 mytest:alpha
#docker run -it --env-file=env.ini -p9292:9292 mytest:alpha /bin/sh