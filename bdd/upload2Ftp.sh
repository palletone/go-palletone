#!/bin/bash

HOST=39.105.191.26
USER=$1
PASS=$2
LCD=$3
RCD=pub

lftp -u $USER,$PASS $HOST << EOF
cd $RCD
put $LCD
EOF