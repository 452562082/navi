#!/bin/sh

ZOOKEEPER_HOSTS=$(echo $ZK_HOSTS)

SERVER_HOSTS=$(echo $SERVER_HOSTS)

echo ZOOKEEPER_HOSTS = ${ZOOKEEPER_HOSTS}


update_zookeeper_hosts() {

    if [ "${ZOOKEEPER_HOSTS}" != "" ]; then
        sed -i 's/\("zookeeper_hosts": "\).*/\1'"${ZOOKEEPER_HOSTS}"'",/g' cfg.json
        echo "update zookeeper_hosts to ${ZOOKEEPER_HOSTS} success"
		return 0
    else
        echo "environment variable 'ZK_HOSTS' is not set"
	    return 1
    fi
}

update_server_hosts() {

    if [ "${SERVER_HOSTS}" != "" ];then
        sed -i 's/\("server_hosts": "\).*/\1'"${SERVER_HOSTS}"'",/g' cfg.json
        echo "update server_hosts_hosts to ${SERVER_HOSTS} success"
		return 0
    else
         echo "environment variable 'SERVER_HOSTS' is not set"
	    return 1
    fi
}

main() {
    update_zookeeper_hosts
   	local ret=$?
	if [ $ret -eq 1 ]; then
       exit 1
	fi

	update_server_hosts
   	local ret=$?
	if [ $ret -eq 1 ]; then
       exit 1
	fi
}

main
exit $?