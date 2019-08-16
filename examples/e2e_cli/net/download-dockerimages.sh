#!/bin/bash -eu
#
# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#


##################################################
# This script pulls docker images from palletone
# docker hub repository and Tag it as
# palletone/mediator-<image> latest tag
##################################################

#Set ARCH variable i.e ppc64le,s390x,x86_64,i386
dockerPalletOnePull() {
  local PALLETONE_TAG=$1
  count=1;
  while [ $count -le 5 ];
  do
      echo "==> PALLETONE IMAGE: palletone/mediator$count"
      echo
      docker pull palletone/mediator$count:$PALLETONE_TAG
      docker tag palletone/mediator$count:$PALLETONE_TAG palletone/mediator$count
      let ++count
  done
}

usage() {
      echo "Description "
      echo
      echo "Pulls docker images from palletone dockerhub repository"
      echo "tag as palletone/mediator<image>:latest"
      echo
      echo "USAGE: "
      echo
      echo "./download-dockerimages.sh tag"
      echo
      echo
      echo "EXAMPLE:"
      echo "./download-dockerimages.sh 1.0.1"
      echo
      echo "By default, pulls mediator 1.0.1 docker images"
      echo "from palletone dockerhub"
}

usage

echo "===> Pulling mediator Images"
dockerPalletOnePull $1

echo "==> PALLETONE IMAGE: palletone/normalnode"
echo 
docker pull palletone/normalnode:$1
docker tag palletone/normalnode:$1 palletone/normalnode

echo
echo "===> List out palletone docker images"
docker images | grep palletone*

