#!/bin/bash

go install -v
if [!-n "$1"]; then
    echo "ServiceName为空"
    exit 1
fi

serviceName = $1
navi_builder create $serviceName

: << !
dockerfilePath = $2
#docker build
!