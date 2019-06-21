#!/bin/bash

HOST=39.105.191.26
USER=$1
PASS=$2
LCD=$3
RCD=$4
RNAME=$5
echo "script start at `date "+%Y-%m-%d %H:%M:%S"`"
lftp -u $USER,$PASS $HOST << EOF
echo "------"
#cd pub
#ls
#echo "1111111"
#mkdir $RCD
#cd $RCD
#ls
#echo "222222"
put $LCD $RNAME
echo "done put"
bye
EOF
echo "script end at `date "+%Y-%m-%d %H:%M:%S"`"
