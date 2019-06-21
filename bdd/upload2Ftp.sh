#!/bin/bash

HOST=39.105.191.26
USER=$1
PASS=$2
LCD=$3
RCD=$4
RNAME=$5
echo "script start at `date "+%Y-%m-%d %H:%M:%S"`"
lftp -e "cd pub; mkdir $RCD; cd $RCD; put $LCD -o $RNAME;exit" -u $USER,$PASS $HOST
echo "script end at `date "+%Y-%m-%d %H:%M:%S"`"