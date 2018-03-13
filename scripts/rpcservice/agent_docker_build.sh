#!/bin/sh

mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/rpcservice/rpc/logs
mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/rpcservice/rpc/libs
mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/rpcservice/rpc/bin
mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/rpcservice/rpc/etc

# 编译rpc_server MyTest
chmod 777 $GOPATH/src/kuaishangtong/navi/cmd/navi-agent/server_example/rpc_server/run.sh
cd $GOPATH/src/kuaishangtong/navi/cmd/navi-agent/server_example/rpc_server
sh -c ./run.sh

# 将 MyTest 拷贝到 rpc/bin 下
cp -a ./mytest $GOPATH/src/kuaishangtong/navi/scripts/rpcservice/rpc/bin

# 编译 navi-agent
cd  $GOPATH/src/kuaishangtong/navi/cmd/navi-agent
go build -v

cd $GOPATH/src/kuaishangtong/navi/scripts/rpcservice

# 将 脚本，navi-agent 及配置文件拷贝到 rpc 下
cp -a $GOPATH/src/kuaishangtong/navi/scripts/update_config.sh rpc/

cp -a $GOPATH/src/kuaishangtong/navi/scripts/rpcservice/run.sh rpc/

cp -a $GOPATH/src/kuaishangtong/navi/cmd/navi-agent/navi-agent rpc/bin

cp -a $GOPATH/src/kuaishangtong/navi/cmd/navi-agent/cfg.json rpc/etc

# 将 MyTest 依赖库拷贝到 rpc/libs 下
cp -a /usr/local/lib/libthriftnb-0.10.0.so rpc/libs
cp -a /usr/lib64/libevent-2.0.so.5 rpc/libs
cp -a /usr/lib64/libevent-2.0.so.5.1.9 rpc/libs
cp -a /usr/local/lib/libthrift-0.10.0.so rpc/libs

echo "FROM centos:7
COPY ./rpc /rpc
RUN chmod 777 /rpc/*.sh
CMD [\"/rpc/run.sh\"]
" > Dockerfile

docker build -t mytest:alpha .