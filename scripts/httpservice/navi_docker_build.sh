#!/bin/sh

ServiceName="MyTest"

mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/httpservice/rpc/logs
mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/httpservice/rpc/bin
mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/httpservice/rpc/etc

#编译安装 navi-cli
cd $GOPATH/src/kuaishangtong/navi/cmd/navi-cli/navi_builder
go build -v
go install

# 编译http_server MyTest
navi_builder create $ServiceName
cd $GOPATH/src/$ServiceName
go build -v

#将MyTest 拷贝到 rpc/bin 下
#cp -a ./MyTest $GOPATH/src/kuaishangtong/navi/scripts/httpservice/rpc/bin

#将脚本及配置文件拷贝到 rpc 下
cp -a $GOPATH/src/kuaishangtong/navi/scripts/update_config.sh rpc/

cp -a $GOPATH/src/kuaishangtong/navi/scripts/httpservice/run.sh rpc/

cp -a $GOPATH/src/kuaishangtong/cmd/navi-agent/navi-agent rpc/bin

cp -a $GOPATH/src/$ServiceName/service.yaml rpc/etc

#删除生成文件
rm -rf $GOPATH/src/$ServiceName

#构建Docker镜像
echo "FROM centos:7
COPY ./rpc /rpc
RUN chmod 777 /rpc/*.sh
" > Dockerfile

docker build -t mytest:alpha .
