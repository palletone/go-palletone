#!/bin/bash

HOST=39.105.191.26
USER=$1
PASS=$2
LCD=$3
RCD=$4
RNAME=$5
echo $USER
echo $PASS
lftp -u $USER,$PASS $HOST << EOF
cd pub
mkdir $RCD
cd $RCD
put $LCD $RNAME
bye
EOF