#!/bin/bash
#######################################################
#                                                     #
# 文件:generateArtifacts.sh                           #
# 作者:dev@palletone.com                              #
# 日期:2018-11-26                                     #
#                                                     #
#######################################################

set -e

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
	#(3.1)创建相应目录
	mkdir -p ${PTNCFGP}/channel-artifacts/mediator${NUMNODE}

	#(3.2)创建toml配置文件
	generateToml ${GPTNRBIN} mediator${NUMNODE} ${PTNCFGP}
	
	#(3.3)针对第0个节点创建创世区块文件
	if [ ${NUMNODE} -eq 0 ]; then
		generateGenesis ${GPTNRBIN} mediator${NUMNODE} ${PTNCFGP}
	fi

	#(3.4)修改toml配置文件信息
	##toml文件路径
        TOML_LOC=${PTNCFGP}/channel-artifacts/mediator${NUMNODE}/ptn-config.toml
	##json文件路径
        JSON_LOC=${PTNCFGP}/channel-artifacts/mediator${NUMNODE}/ptn-genesis.json
        JSON_TMP=${PTNCFGP}/channel-artifacts/mediator0/ptn-genesis.json
	##得到账号信息
	account=`getPtnAccount ${PWD}`
	##得到秘钥相关
	keys=`getPtnKeys ${GPTNRBIN}`
        prvatekey=`echo ${keys} | cut -d \_ -f 1`
        publickey=`echo ${keys} | cut -d \_ -f 2`
	##节点相关
	nodeinfo=`getPtnNode ${GPTNRBIN}`
	##更新toml和json
	updateTomlFile ${TOML_LOC} ${account} ${prvatekey} ${publickey}
	##将所有的节点信息同步到第一个节点中的创世文件中
	echo ${nodeinfo}
	updateGenesis  ${JSON_TMP} ${account} ${publickey} ${nodeinfo}
	
	#(3.5)将数据目录进行移动到相应目录
	echo "开始拷节点 ${NUMNODE} 目录"
	mv -f ${PWD}/palletone ${PWD}/log ${PTNCFGP}/channel-artifacts/mediator${NUMNODE}
	echo "##########################################################"
	echo "生成mediator${NUMNODE}目录信息"
	echo "##########################################################"
	rm -f ${PWD}/ptn-config.toml
	
	#(3.6)更新节点
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
ExecInit ${PTNCFGP}/channel-artifacts ${NUMNODES}

#(7)退出
exit 0
