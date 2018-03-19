#!/bin/sh

mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/gatewayservice/gateway/logs
mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/gatewayservice/gateway/libs
mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/gatewayservice/gateway/bin
mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/gatewayservice/gateway/bin/conf

# 编译 gateway
cd $GOPATH/src/kuaishangtong/navi/gateway
go build -v

cd $GOPATH/src/kuaishangtong/navi/scripts/gatewayservice

# 将 脚本，gateway 及配置文件拷贝到 gateway 下
cp -a $GOPATH/src/kuaishangtong/navi/scripts/update_config.sh gateway/
chmod 777 gateway/update_config.sh

cp -a $GOPATH/src/kuaishangtong/navi/scripts/gatewayservice/run.sh gateway/
chmod 777 gateway/run.sh

cp -a $GOPATH/src/kuaishangtong/navi/gateway/gateway gateway/bin

cp -a $GOPATH/src/kuaishangtong/navi/gateway/conf/app.conf gateway/bin/conf

echo "FROM centos:7
COPY gateway /gateway
CMD [\"/gateway/run.sh\"]
" > Dockerfile
#
docker build -t mytest:alpha .

#rm -rf ./gateway