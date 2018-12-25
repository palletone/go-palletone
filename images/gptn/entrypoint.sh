#!/bin/bash

#初始化Dockerfile.in里面的启动参数 
function initDock() {
	#得到配置文件名称和创世区块文件
	FILE_TOML=$1
	FILE_GENESIS=$2

	#判断路径下面是否包含创世区块文件,如果包含创世区块文件，则认为新建一条链,否则同步主链或测试链
	if [ -f "${FILE_GENESIS}" ]; then
		echo
        	echo "##########################################################"
		echo "#####             Init genesis file              #########"
        	echo "##########################################################"
		echo
		gptn --datadir=/var/palletone/production init ${FILE_GENESIS}
		sleep 2
	fi
	
	#启动主程序
	echo
        echo "##########################################################"
        echo "#####           start palletone gptn             #########"
        echo "##########################################################"
        echo
	gptn --datadir=/var/palletone/production  --configfile=${FILE_TOML}

}

#############################################################################################
#若是一个参数,则直接启动主程序;若是两个参数，则先初始区块文件，再启动主程序；否则，提示错误 #
#############################################################################################
#(1)接收相关参数
TOML=$1
GENS=$2
#打印提示信息
echo "Receiving parameters are : ConfigurationFile => ${TOML} ====== GenesisFile => ${GENS}"

#(2)确定必须存在配置文件
if [ ! -f ${TOML} ]; then
	echo "Error: Toml configuration file does not exist !"
	exit 1
fi

#(3)执行逻辑
if [ $# -eq 1 -o $# -eq 2 ]; then
	initDock ${TOML} ${GENS}
else
	echo "Error: Relevant parameters are: ${TOML} - ${GENS}"
	exit 1
fi
