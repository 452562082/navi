#!/bin/sh

ServiceName="MyTest"

mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/httpservice/navi/logs
mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/httpservice/navi/bin
mkdir -p $GOPATH/src/kuaishangtong/navi/scripts/httpservice/navi/etc

#编译安装 navi-cli
cd $GOPATH/src/kuaishangtong/navi/cmd/navi-cli/navi_builder
go build -v

# 编译http_server MyTest
navi_builder create $ServiceName
cd $GOPATH/src/$ServiceName
go build -v

cd $GOPATH/src/kuaishangtong/navi/scripts/httpservice

#将脚本及配置文件拷贝到 navi 下
cp -a $GOPATH/src/kuaishangtong/navi/scripts/update_config.sh navi/
chmod 777 navi/update_config.sh

cp -a $GOPATH/src/kuaishangtong/navi/scripts/httpservice/run.sh navi/
chmod 777 navi/run.sh

cp -a $GOPATH/src/$ServiceName/$ServiceName navi/bin

cp -a $GOPATH/src/$ServiceName/service.yaml navi/etc

#删除生成文件
rm -rf $GOPATH/src/$ServiceName

#构建Docker镜像
echo "FROM centos:7
COPY ./navi /navi
#CMD [\"/navi/run.sh\"]
" > Dockerfile

docker build -t mytest_http:alpha .
