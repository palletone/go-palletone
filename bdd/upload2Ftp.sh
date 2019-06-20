#!/bin/bash

HOST=39.105.191.26
USER=$1
PASS=$2
LCD=$3
RCD=pub
echo $USER
echo $PASS
echo $LCD
lftp -u $USER,$PASS $HOST << EOF
echo "this is a test"
cd $RCD
put $LCD
EOF