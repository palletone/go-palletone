#!/bin/bash

function StartGPTN()
{
    echo ===============$1===============
    if [ $1 -ne 1 ] ;then
        nohup ./gptn --datadir node$1/palletone --configfile node$1/ptn-config.toml >> node$1/nohup.out &
    else
        nohup ./gptn --datadir node$1/palletone --configfile node$1/ptn-config.toml --noProduce --staleProduce  >> node$1/nohup.out &
    fi
}


function LoopStart()  
{  
    count=1;  
    while [ $count -le $1 ] ;  
    do  
    StartGPTN $count 
    let ++count;  
    done 
    nohup ./gptn --datadir node_test4/palletone --configfile node_test4/ptn-config.toml >> node_test4/nohup.out &
    #nohup ./gptn --datadir node_test5/palletone --configfile node_test5/ptn-config.toml >> node_test5/nohup.out &
    #nohup ./gptn --datadir node_test6/palletone --configfile node_test6/ptn-config.toml >> node_test6/nohup.out &
    #nohup ./gptn --datadir node_test7/palletone --configfile node_test7/ptn-config.toml >> node_test7/nohup.out & 
    return 0;  
}

n=3
#if [ -n "$1" ]; then
#    n=$1
#else
#    read -p "Please input the numbers of nodes you want: " n;
#fi
 
LoopStart $n;
