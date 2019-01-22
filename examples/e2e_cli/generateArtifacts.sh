#!/bin/bash
#######################################################
#                                                     #
# 文件:generateArtifacts.sh                           #
# 作者:dev@palletone.com                              #
# 日期:2018-11-26                                     #
#                                                     #
#######################################################

#set -e

#引用公共脚本
source ${PWD}/scripts/getCommon.sh

#创建相关环境变量
export PTNROOT=${PWD}/../..
export PTNCFGP=${PWD}
export GPTNROOT=${PTNROOT}/build/bin
export GPTNRBIN=${GPTNROOT}/gptn

#######################################################
#主程序逻辑
#######################################################
#(0)判断参数是否大于1，否则提示错误
if [ $# -gt 1 ]; then
	echo "Usage: ./generateArtifacts.sh <numNod>"
	exit 1
elif [ $# -eq 1 ]; then
	NUMNODES=${1}
	#判断是否为奇数
	if [ `expr ${NUMNODES} % 2` -eq 0 ];then
		echo "Error: The number of mediator nodes cannot be even!"
		exit 1
	fi
else
	#将信息设置为1个节点和Local方式
	NUMNODES=1
fi

#(1)判断是否存在channel-artifacts目录
if [ -d "${PWD}/channel-artifacts" ]; then
	echo "Error: the directory already exists,please re-check."
	exit 1
fi

#(2)创建父目录
generateChannelArtifacts ${PWD}

#(3)循环创建相应目录及其配置文件
NUMNODE=0
while [ ${NUMNODE} -lt ${NUMNODES} ]
do
	#(3.1)创建相应目录,并拷贝可执行文件
	if [ ! -d "${PTNCFGP}/channel-artifacts/mediator${NUMNODE}" ];then
		##创建目录
		mkdir -p ${PTNCFGP}/channel-artifacts/mediator${NUMNODE}
		##拷贝文件
		cp -rf ${GPTNRBIN} ${PTNCFGP}/channel-artifacts/mediator${NUMNODE}
		##进入目录
		cd ${PTNCFGP}/channel-artifacts/mediator${NUMNODE}
	fi

	#(3.2)相关变量信息
	##当前临时目录
	TEMPPATH=${PTNCFGP}/channel-artifacts/mediator${NUMNODE}
	##使用当前目录下面的gptn
        TMPGPTNBIN=${PTNCFGP}/channel-artifacts/mediator${NUMNODE}/gptn
	##TOML相关文件路径
	TOML_LOC=${PTNCFGP}/channel-artifacts/mediator${NUMNODE}/ptn-config.toml
	##JSON文件路径
        JSON_LOC=${PTNCFGP}/channel-artifacts/mediator${NUMNODE}/ptn-genesis.json
        JSON_TMP=${PTNCFGP}/channel-artifacts/mediator0/ptn-genesis.json

	#(3.3)创建toml配置文件
	generateToml ${TMPGPTNBIN} mediator${NUMNODE} ${TEMPPATH}
	
	#(3.4)针对第0个节点创建创世区块文件
	if [ ${NUMNODE} -eq 0 ]; then
		generateGenesis ${TMPGPTNBIN} mediator${NUMNODE} ${TEMPPATH}
	fi

	#(3.4)修改toml配置文件信息
	##得到账号信息
	account=`getPtnAccount ${PTNCFGP}  ${TEMPPATH}`
	##得到秘钥相关
	keys=`getPtnKeys ${TMPGPTNBIN} ${TEMPPATH}`
        prvatekey=`echo ${keys} | cut -d \_ -f 1`
        publickey=`echo ${keys} | cut -d \_ -f 2`
	echo "privatekey=> $prvatekey === publickey=> $publickey"
	##节点相关
	nodeinfo=`getPtnNode ${TMPGPTNBIN} ${TEMPPATH}`
	##更新toml和json
	updateTomlFile ${TOML_LOC} ${account} ${prvatekey} ${publickey}
	##将所有的节点信息同步到第一个节点中的创世文件中
	updateGenesis  ${JSON_TMP} ${account} ${publickey} ${nodeinfo}
	
	#(3.5)更新节点
	let ++NUMNODE
done


#(4)针对第一个节点的创世区块文件进行更新活跃mediator信息
##第一个节点的创世区块文件
JSON_LOC=${PTNCFGP}/channel-artifacts/mediator0/ptn-genesis.json
##更新
replaceJson ${JSON_LOC}

#(5)添加pnode信息
modifyStaticNodes ${JSON_LOC} ${NUMNODES} ${PTNCFGP}/channel-artifacts

#(6)初始化leveldb
#ExecInit ${PTNCFGP}/channel-artifacts ${NUMNODES}

#(7)退出
exit 0
