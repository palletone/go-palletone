#!/bin/bash

function StartGPTN()
{
    echo ===============$1===============
    if [ $1 -ne 1 ] ;then
        nohup ./gptn --datadir node$1/palletone --configfile node$1/ptn-config.toml >> node$1/nohup.out &
    else
        nohup ./gptn --datadir node$1/palletone --configfile node$1/ptn-config.toml --noProduce --staleProduce  >> node$1/nohup.out &
        #nohup ./gptn --datadir node$1/palletone --configfile node$1/ptn-config.toml --noProduce --staleProduce --allowConsecutive  >> node$1/nohup.out &
    fi
}


function LoopStart()  
{  
    count=1;  
    while [ $count -le $1 ] ;  
    do  
    StartGPTN $count 
    sleep 3
    let ++count;  
    done  
    return 0;  
}

n=
if [ -n "$1" ]; then
    n=$1
else
    read -p "Please input the numbers of nodes you want: " n;
fi
 
LoopStart $n;
