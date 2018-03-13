#!/bin/sh

RED_COLOR='\E[1;31m'  #红
GREEN_COLOR='\E[1;32m' #绿
YELOW_COLOR='\E[1;33m' #黄
BLUE_COLOR='\E[1;34m'  #蓝
PINK='\E[1;35m'      #粉红
RES='\E[0m'

echo_success() {
    local str=$1
    echo -e "${GREEN_COLOR}[SUCC]${RES} ${str}"
}

echo_warning() {
    local str=$1
    echo -e "${YELOW_COLOR}[SUCC]${RES} ${str}"
}

echo_failure() {
    local str=$1
    echo -e "${RED_COLOR}[ERROR]${RES} ${str}"
}

ZOOKEEPER_HOSTS=$(echo $ZK_HOSTS)

SERVER_HOSTS=$(echo $SERVER_HOSTS)

echo_warning ZOOKEEPER_HOSTS = ${ZOOKEEPER_HOSTS}

echo_warning SERVER_HOSTS = ${SERVER_HOSTS}


update_zookeeper_hosts() {

    if [ "${ZOOKEEPER_HOSTS}" != "" ]; then
        sed -i 's/\("zookeeper_hosts": "\).*/\1'"${ZOOKEEPER_HOSTS}"'",/g' cfg.json
        echo_success "update zookeeper_hosts to ${ZOOKEEPER_HOSTS}"
		return 0
    else
        echo_failure "environment variable 'ZK_HOSTS' is not set"
	    return 1
    fi
}

update_server_hosts() {

    if [ "${SERVER_HOSTS}" != "" ];then
        sed -i 's/\("server_hosts": "\).*/\1'"${SERVER_HOSTS}"'",/g' cfg.json
        echo_success "update server_hosts_hosts to ${SERVER_HOSTS}"
		return 0
    else
        echo_failure "environment variable 'SERVER_HOSTS' is not set"
	    return 1
    fi
}

main() {
    update_zookeeper_hosts
   	local ret=$?
	if [ $ret -eq 1 ]; then
        return 2
	fi

	update_server_hosts
   	local ret=$?
	if [ $ret -eq 1 ]; then
        return 2
	fi
}

main
exit $?