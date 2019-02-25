#!/bin/bash

function StartGPTN()
{
    echo ===============$1===============
    if [ $1 -ne 1 ] ;then
        nohup ./gptn --datadi node$1/palletone --configfile node$1/ptn-config.toml >> node$1/nohup.out &
    else
        nohup ./gptn --datadi node$1/palletone --configfile node$1/ptn-config.toml --noProduce --staleProduce  >> node$1/nohup.out &
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
    return 0;  
}
read -p "Please input the numbers of nodes you want: " n;  
 
LoopStart $n;
