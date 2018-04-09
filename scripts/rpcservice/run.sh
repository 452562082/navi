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
    /rpc/update_config.sh agent /rpc/etc/cfg.json
}

start_rpc_server() {
    export LD_LIBRARY_PATH=/usr/local/lib:/rpc/libs
    echo_success "start mytest"
    nohup /rpc/bin/mytest 9292 >/rpc/logs/mytest.log 2>&1 &
}

start_agent() {
    echo_success "start navi-agent"
    /rpc/bin/navi-agent -c /rpc/etc/cfg.json
}

start () {
    update_config

    if [ $? != "0" ]; then
        echo_failure "update_config failed"
        exit 2
    fi

    start_rpc_server

    sleep 3

    start_agent

}

#start
#exit 0

case "$1" in
	start)
		start
		;;
	restart)
		start_rpc_server
		;;
	*)
		echo -e "${YELOW_COLOR}Usage: $0 {start|restart} ${RES}"
		exit 1
esac

exit $RETVAL