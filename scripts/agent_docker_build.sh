#!/bin/sh

mkdir -p rpc/logs
mkdir -p rpc/libs
mkdir -p rpc/bin
mkdir -p rpc/etc


chmod 777 ../cmd/navi-agent/server_example/rpc_server/run.sh
cd ../cmd/navi-agent/server_example/rpc_server
sh -c ./run.sh

cp -a ./mytest $GOPATH/src/kuaishangtong/navi/scripts/rpc/bin

cd  $GOPATH/src/kuaishangtong/navi/cmd/navi-agent

go build -v

cd ../../scripts

cp -a ./update_config.sh rpc/

cp -a ./run.sh rpc/

cp -a ../cmd/navi-agent/navi-agent rpc/bin

cp -a ../cmd/navi-agent/cfg.json rpc/etc

cp -a /usr/local/lib/libthriftnb-0.10.0.so rpc/libs
cp -a /usr/lib64/libevent-2.0.so.5 rpc/libs
cp -a /usr/lib64/libevent-2.0.so.5.1.9 rpc/libs
cp -a //usr/local/lib/libthrift-0.10.0.so rpc/libs



echo "FROM centos:7
COPY ./rpc /rpc
RUN chmod 777 /rpc/*.sh
ENTRYPOINT [\"/rpc/update_config.sh\", \"agent\", \"/rpc/etc/cfg.json\"]
#CMD [\"/rpc/run.sh\"]
" > Dockerfile

docker build -t mytest:alpha .