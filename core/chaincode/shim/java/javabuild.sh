#!/bin/bash

#
#Copyright DTCC 2016 All Rights Reserved.
#
#Licensed under the Apache License, Version 2.0 (the "License");
#you may not use this file except in compliance with the License.
#You may obtain a copy of the License at
#
#         http://www.apache.org/licenses/LICENSE-2.0
#
#Unless required by applicable law or agreed to in writing, software
#distributed under the License is distributed on an "AS IS" BASIS,
#WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#See the License for the specific language governing permissions and
#limitations under the License.
#
#
set -e
PARENTDIR=$(pwd)
ARCH=`uname -m`

function getProxyHost {
  ADDR=${1#*://}
  echo ${ADDR%:*}
}

function getProxyPort {
  ADDR=${1#*://}
  echo ${ADDR#*:}
}

[ -n "$http_proxy" ] && JAVA_OPTS="$JAVA_OPTS -Dhttp.proxyHost=$(getProxyHost $http_proxy) -Dhttp.proxyPort=$(getProxyPort $http_proxy)"
[ -n "$https_proxy" ] && JAVA_OPTS="$JAVA_OPTS -Dhttps.proxyHost=$(getProxyHost $https_proxy) -Dhttps.proxyPort=$(getProxyPort $https_proxy)"
[ -n "$HTTP_PROXY" ] && JAVA_OPTS="$JAVA_OPTS -Dhttp.proxyHost=$(getProxyHost $HTTP_PROXY) -Dhttp.proxyPort=$(getProxyPort $HTTP_PROXY)"
[ -n "$HTTPS_PROXY" ] && JAVA_OPTS="$JAVA_OPTS -Dhttps.proxyHost=$(getProxyHost $HTTPS_PROXY) -Dhttps.proxyPort=$(getProxyPort $HTTPS_PROXY)"
export JAVA_OPTS

if [ x$ARCH == xx86_64 ]
then
    gradle -q -b ${PARENTDIR}/core/chaincode/shim/java/build.gradle clean
    gradle -q -b ${PARENTDIR}/core/chaincode/shim/java/build.gradle build
    cp -r ${PARENTDIR}/core/chaincode/shim/java/build/libs /root/
else
    echo "FIXME: Java Shim code needs work on ppc64le and s390x."
    echo "Commenting it for now."
fi
