#!/bin/sh

mkdir mytest

chmod 777 ../cmd/navi-agent/server_example/rpc_server/run.sh
sh -c ../cmd/navi-agent/server_example/rpc_server/run.sh

cp -a ../cmd/navi-agent/server_example/rpc_server/mytest mytest/

cd ../cmd/navi-agent

go build -v

cd ../../scripts

cp -a ./update_config.sh mytest/

cp -a ../cmd/navi-agent/navi-agent mytest/

cp -a ../cmd/navi-agent/cfg.json mytest/

echo "FROM busybox
COPY ./mytest /go/src/mytest
ENTRYPOINT [\"sh\", \"-c\", \"/go/src/mytest/update_config.sh\"]
CMD [\"sh\", \"-c\", \"/go/src/mytest/run.sh\"]
" > Dockerfile
