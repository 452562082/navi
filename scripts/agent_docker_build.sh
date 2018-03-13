#!/bin/sh

mkdir mytest

chmod 777 ../cmd/navi-agent/server_example/rpc_server/run.sh
cd ../cmd/navi-agent/server_example/rpc_server
sh -c ./run.sh

cp -a ./mytest $GOPATH/src/kuaishangtong/navi/scripts/mytest/

cd  $GOPATH/src/kuaishangtong/navi/cmd/navi-agent

go build -v

cd ../../scripts

cp -a ./update_config.sh mytest/

cp -a ../cmd/navi-agent/navi-agent mytest/

cp -a ../cmd/navi-agent/cfg.json mytest/

echo "FROM busybox
COPY ./mytest /go/src/mytest
#ENTRYPOINT [\"sh\", \"-c\", \"/go/src/mytest/update_config.sh\", \"agent\", \"/go/src/mytest/cfg.json\"]
#CMD [\"sh\", \"-c\", \"/go/src/mytest/run.sh\"]
" > Dockerfile

docker build -t mytest:alpha .