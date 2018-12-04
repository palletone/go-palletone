#!/bin/bash

source ./modifyconfig.sh

function ExecInit()
{
    
    count=1  
    while [ $count -le $1 ] ;  
    do  
    #echo $count;
    #ExecDeploy $count 
    if [ $count -eq 1 ] ;then
    cd node$count
    cp ../init.sh .
    gptninit=`./init.sh`
    `echo $gptninit`
    path=`pwd`
    fullpath=${path}"/palletone/gptn/leveldb"
    echo "leveldb path:"$fullpath
    if [ ! -d $fullpath ]; then
        echo "====================init err=================="
        return
    fi
        rm -rf init.sh log
        cd ../
    else
    echo $count
        cd node$count
        cp ../node1/palletone/gptn/leveldb ./palletone/gptn/. -rf
        rm -rf log
        cd ../
    fi
    let ++count;  
    sleep 1;  
    done 
    echo "====================init ok====================="
    return 0;  
    
}



function replacejson()
{
    length=`cat $1 |jq '.initialMediatorCandidates| length'`

    MinMediatorCount="MinMediatorCount"
    line=`awk "/$MinMediatorCount/{print NR}" $1`
    newMinMediatorCount="\"MinMediatorCount\":$length,"
    replace=`sed -e "${line}c $newMinMediatorCount" $1`
    rm $1
    `echo $replace >>t.json`
    jq -r . t.json >> $1
    rm t.json


    add=`cat $1 | 
       jq "to_entries | 
       map(if .key == \"initialActiveMediators\" 
          then . + {\"value\":$length} 
          else . 
          end
         ) | 
      from_entries"`

    rm $1
    echo $add >> temp.json
    jq -r . temp.json >> $1
    rm temp.json
    echo "======modify genesis json ok======="
}

function ExecDeploy()
{
    echo ===============$1===============
    mkdir "node"$1
    #if [ $1 -eq 4 ] ;then
    #echo "=="$1
    cp gptn ./createaccount.sh modifyconfig.sh modifyjson.sh init.sh node$1
    cd node$1
    /bin/bash modifyconfig.sh
    #source ./modifyjson.sh
    source ./modifyconfig.sh
    ModifyConfig $1
    rm *.sh
    cd ../
    #else
    #echo $1
    #fi
}



function LoopDeploy()  
{  
    count=1;  
    while [ $count -le $1 ] ;  
    do  
    #echo $count;
    ExecDeploy $count 
    let ++count;  
    sleep 1;  
    done  
    return 0;  
}
path=`echo $GOPATH`
src=/src/github.com/palletone/go-palletone/build/bin/gptn
fullpath=$path$src
cp $fullpath .
read -p "Please input the numbers of nodes you want: " n;  
 
LoopDeploy $n;

json="node1/ptn-genesis.json"
replacejson $json 


ModifyStaticNodes $n

ExecInit $n


