#!/bin/bash

#
# Copyright (c) 2023 OceanBase
# OBAgent is licensed under Mulan PSL v2.
# You can use this software according to the terms and conditions of the Mulan PSL v2.
# You may obtain a copy of Mulan PSL v2 at:
#          http://license.coscl.org.cn/MulanPSL2
# THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
# EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
# MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
# See the Mulan PSL v2 for more details.
#

PROJECT_DIR=$1
PROJECT_NAME=$2
VERSION=$3
RELEASE=$4

CUR_DIR=$(dirname $(readlink -f "$0"))
TOP_DIR=~/.rpm_build
echo "[BUILD] args: CURDIR=${CUR_DIR} PROJECT_NAME=${PROJECT_NAME} VERSION=${VERSION} RELEASE=${RELEASE}"

# prepare rpm build dirs
rm -rf $TOP_DIR
mkdir -p $TOP_DIR/BUILD $TOP_DIR/RPMS $TOP_DIR/SOURCES $TOP_DIR/SPECS $TOP_DIR/SRPMS
echo "dir is: $(ls -l ${TOP_DIR})"

# build rpm
cd $CUR_DIR
export PROJECT_NAME=${PROJECT_NAME}
export VERSION=${VERSION}
export RELEASE=${RELEASE}
rpmbuild --define "_topdir $TOP_DIR" -bb $PROJECT_NAME.spec
find $TOP_DIR/ -name "*.rpm" -exec mv {} . 2>/dev/null \;
echo "RPM path: $(find . -name '*.rpm')"
