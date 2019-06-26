#!/bin/bash
rm -rf node* gptn
./deploy.sh
sleep 1


info=`cat node_test4/ptn-config.toml |grep 'StaticNodes='|awk '{print $1}'`
arrs=`echo $info`
arr=`echo ${arrs:12}`
node1=`echo "${arr}" | jq -c '.[1]'`
node2=`echo "${arr}" | jq -c '.[2]'`

newarrStaticNodes="StaticNodes=[$node1]"
sed -i '/^StaticNodes/c'$newarrStaticNodes'' node_test4/ptn-config.toml

cd node_test4 
while :
do
info=`./gptn nodeInfo`
tempinfo=`echo $info | sed -n '$p'| awk '{print $NF}'`
length=`echo ${#tempinfo}`
nodeinfotemp=${tempinfo:0:$length}
nodeinfo=`echo ${nodeinfotemp//^M/}`
length=`echo ${#nodeinfo}`
b=140
if [ "$length" -lt "$b" ]
then
    continue
else
    break
fi
done

cd ../

echo $nodeinfo
newarrStaticNodes="StaticNodes=[\"$nodeinfo\"]"
sed -i '/^StaticNodes/c'$newarrStaticNodes'' node_test5/ptn-config.toml

newSyncMode="SyncMode=\"light\""
sed -i '/^SyncMode/c'$newSyncMode'' node_test5/ptn-config.toml


newarrStaticNodes="StaticNodes=[\"$nodeinfo\"]"
sed -i '/^StaticNodes/c'$newarrStaticNodes'' node_test6/ptn-config.toml

newSyncMode="SyncMode=\"light\""
sed -i '/^SyncMode/c'$newSyncMode'' node_test6/ptn-config.toml


newarrStaticNodes="StaticNodes=[$node2]"
sed -i '/^StaticNodes/c'$newarrStaticNodes'' node_test7/ptn-config.toml

newSyncMode="SyncMode=\"light\""
sed -i '/^SyncMode/c'$newSyncMode'' node_test7/ptn-config.toml

sleep 1
./start.sh

sleep 1
numcommand=`ps -ef|grep gptn |wc -l`
num=`echo $numcommand | sed -n '$p'| awk '{print $NF}'`

if [ $num -eq 8 ];then
    echo "============preset start ok num"$num"============"
else
    echo "============preset start err num:"$num"============"
fi
