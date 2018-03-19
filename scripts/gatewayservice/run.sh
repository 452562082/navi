#!/bin/sh

RED_COLOR='\E[1;31m'  #红
GREEN_COLOR='\E[1;32m' #绿
YELOW_COLOR='\E[1;33m' #黄
BLUE_COLOR='\E[1;34m'  #蓝
PINK='\E[1;35m'      #粉红
RES='\E[0m'
RETVAL=0


echo_success() {
    local str=$1
    echo -e "${GREEN_COLOR}[SUCC]${RES} ${str}"
}

echo_warning() {
    local str=$1
    echo -e "${YELOW_COLOR}[WARN]${RES} ${str}"
}

echo_failure() {
    local str=$1
    echo -e "${RED_COLOR}[ERROR]${RES} ${str}"
}

update_config() {
    /gateway/update_config.sh gateway /gateway/bin/conf/app.conf
}

start_gateway() {
    echo_success "gateway"
    /gateway/bin/gateway
}

main () {
    update_config

    if [ $? != "0" ]; then
        echo_failure "update_config failed"
        exit 2
    fi
    start_gateway
}

main
exit 0