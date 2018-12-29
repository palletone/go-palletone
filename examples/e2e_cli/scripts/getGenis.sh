#!/bin/bash

#(1)执行getInit.sh脚本
gptninit=`./getInit.sh`
`echo $gptninit`
#(2)得到当前目录
path=`pwd`
fullpath=${path}"/palletone/gptn/leveldb"
#(3)检查目录是否存在
echo "leveldb path:"$fullpath
if [ ! -d $fullpath ]; then
	echo "====================init err=================="
	exit 1
fi

echo "====================init OK=================="
exit 0

