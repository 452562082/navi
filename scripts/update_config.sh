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

ZOOKEEPER_HOSTS=$(echo $ZK_HOSTS)

SERVER_HOSTS=$(echo $SERVER_HOSTS)

JAEGER_HOST=$(echo $JAEGER_HOST)

update_json_zookeeper_hosts() {

    echo_warning "ZOOKEEPER_HOSTS = ${ZOOKEEPER_HOSTS}"

    if [ "${ZOOKEEPER_HOSTS}" != "" ]; then
        sed -i 's/\("zookeeper_hosts": "\).*/\1'"${ZOOKEEPER_HOSTS}"'",/g' $1
        echo_success "update zookeeper_hosts to ${ZOOKEEPER_HOSTS}"
		return 0
    else
        echo_failure "environment variable 'ZK_HOSTS' is not set"
	    return 1
    fi
}

update_yaml_zookeeper_hosts() {

    echo_warning "ZOOKEEPER_HOSTS = ${ZOOKEEPER_HOSTS}"

    if [ "${ZOOKEEPER_HOSTS}" != "" ]; then
        #sed -i 's/\("zookeeper_hosts": "\).*/\1'"${ZOOKEEPER_HOSTS}"'",/g' $1
		sed -i 's/\(zookeeper_servers_addr: \)\(.*\)/\1'"${ZOOKEEPER_HOSTS}"'/g' $1
        echo_success "update zookeeper_hosts to ${ZOOKEEPER_HOSTS}"
		return 0
    else
        echo_failure "environment variable 'ZK_HOSTS' is not set"
	    return 1
    fi
}

update_ini_zookeeper_hosts() {

    echo_warning "ZOOKEEPER_HOSTS = ${ZOOKEEPER_HOSTS}"

    if [ "${ZOOKEEPER_HOSTS}" != "" ]; then
		sed -i 's/\(zookeeper.hosts = \)\(.*\)/\1'"${ZOOKEEPER_HOSTS}"'/g' $1
        echo_success "update zookeeper.hosts to ${ZOOKEEPER_HOSTS}"
		return 0
    else
        echo_failure "environment variable 'ZK_HOSTS' is not set"
	    return 1
    fi
}

update_json_server_hosts() {

    echo_warning "SERVER_HOSTS = ${SERVER_HOSTS}"

    if [ "${SERVER_HOSTS}" != "" ];then
        sed -i 's/\("server_hosts": "\).*/\1'"${SERVER_HOSTS}"'",/g' $1
        echo_success "update server_hosts to ${SERVER_HOSTS}"
		return 0
    else
        echo_failure "environment variable 'SERVER_HOSTS' is not set"
	    return 1
    fi
}

update_yaml_server_hosts() {

    echo_warning "SERVER_HOSTS = ${SERVER_HOSTS}"

    if [ "${SERVER_HOSTS}" != "" ];then
        #sed -i 's/\("server_hosts": "\).*/\1'"${SERVER_HOSTS}"'",/g' $1
        sed -i 's/\(http_host: \)\(.*\)/\1'"${SERVER_HOSTS}"'/g' $1
        echo_success "update server_hosts to ${SERVER_HOSTS}"
		return 0
    else
        echo_failure "environment variable 'SERVER_HOSTS' is not set"
	    return 1
    fi
}

update_yaml_file() {

    echo_warning "LOG_FILE = ${LOG_FILE}"

    if [ "${LOG_FILE}" != "" ];then
        sed -i 's/\(file: \)\(.*\)/\1'"${LOG_FILE}"'/g' $1
        echo_success "update file to ${LOG_FILE}"
		return 0
    else
        echo_failure "environment variable 'LOG_FILE' is not set"
	    return 1
    fi
}


update_ini_jaeger_host() {
        echo_warning "JAEGER_HOST = ${JAEGER_HOST}"

    if [ "${JAEGER_HOST}" != "" ]; then
		sed -i 's/\(jaeger.host = \)\(.*\)/\1'"${JAEGER_HOST}"'/g' $1
        echo_success "update jaeger.host to ${JAEGER_HOST}"
		return 0
    else
        echo_failure "environment variable 'JAEGER_HOST' is not set"

    fi
}

update_yaml_jaeger_addr() {

    echo_warning "JAEGER_ADDR = ${JAEGER_ADDR}"

    if [ "${JAEGER_ADDR}" != "" ];then
        sed -i 's/\(jaeger_addr: \)\(.*\)/\1'"${JAEGER_ADDR}"'/g' $1
        echo_success "update file to ${JAEGER_ADDR}"
		return 0
    else
        echo_failure "environment variable 'JAEGER_ADDR' is not set"

	    return 1
    fi
}

print_help() {
    echo -e "${YELOW_COLOR}Usage: $0 {agent|gateway|navi} [config_file] ${RES}"
}

agent() {

    if [ "$1" == "" ];then
       echo_failure "agent config file does not be designated"
       print_help
       exit 2
    fi

    if [ "$1" == "-h" ];then
       print_help
       exit 0
    fi

    update_json_zookeeper_hosts $1
   	local ret=$?
	if [ $ret -eq 1 ]; then
        exit 2
	fi

	update_json_server_hosts $1
   	local ret=$?
	if [ $ret -eq 1 ]; then
        exit 2
	fi
}

navi() {
    if [ "$1" == "" ];then
       echo_failure "navi service.yaml config file does not be designated"
       print_help
       exit 2
    fi

    if [ "$1" == "-h" ];then
       print_help
       exit 0
    fi

    update_yaml_zookeeper_hosts $1
   	local ret=$?
	if [ $ret -eq 1 ]; then
        exit 2
	fi

	update_yaml_server_hosts $1
   	local ret=$?
	if [ $ret -eq 1 ]; then
        exit 2
	fi

	update_yaml_file $1
	local ret=$?
	if [ $ret -eq 1 ]; then
	    exit 2
	fi

	update_yaml_jaeger_addr $1
	local ret=$?
	if [ $ret -eq 1 ]; then
	    exit 2
	fi
}

gateway() {
        if [ "$1" == "" ];then
       echo_failure "gateway config file does not be designated"
       print_help
       exit 2
    fi

    if [ "$1" == "-h" ];then
       print_help
       exit 0
    fi

    update_ini_zookeeper_hosts $1
   	local ret=$?
	if [ $ret -eq 1 ]; then
        exit 2
	fi

	update_ini_jaeger_host $1
   	local ret=$?
	if [ $ret -eq 1 ]; then
        exit 2
	fi
}

case "$1" in
	agent)
		agent $2
		;;
	gateway)
		gateway $2
		;;
	navi)
		navi $2
		;;
	*)
		echo -e "${YELOW_COLOR}Usage: $0 {agent|gateway|navi} [config_file] ${RES}"
		exit 1
esac

exit $RETVAL
