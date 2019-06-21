#!/bin/bash

HOST=39.105.191.26
USER=$1
PASS=$2
LCD=$3
RCD=pub
RNAME=$4
lftp -u $USER,$PASS $HOST << EOF
cd $RCD
#mkdir $RCD
#cd $RCD
put $LCD $RNAME
echo "done put"
bye
EOF