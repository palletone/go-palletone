#!/bin/bash
#######################################################
#公共函数
#######################################################

#######################################################
#(1)创建配置文件所在目录结构
#######################################################
function generateChannelArtifacts() {
	CFGPATH=${1}
        mkdir -p ${CFGPATH}/channel-artifacts
	return 0
}

#######################################################
#(2)生成创世区块文件
#######################################################
function generateGenesis() {
        GPTNGEN=${1}
        #节点名称
        NODENAM=${2}
	#配置目录
	PCFGPAH=${3}

        if [ -f "$GPTNGEN" ]; then
                echo "Using gptn -> $GPTNGEN"
        else
            	echo "Building Genesis File"
		exit 1
        fi

        echo
        echo "##########################################################"
        echo "#####   Generate genesis file using gptn tool    #########"
        echo "##########################################################"
        $GPTNGEN dumpjson
        if [ -f "${PCFGPAH}/ptn-genesis.json" ]; then
                if [ ! -d "${PCFGPAH}/channel-artifacts/${NODENAM}" ]; then
                        mkdir -p ${PCFGPAH}/channel-artifacts/${NODENAM}
                fi

                mv -f ${PCFGPAH}/ptn-genesis.json ${PCFGPAH}/channel-artifacts/${NODENAM}
        else
                echo "Generate genesis file failed,please check!"
        fi
        echo
	
	return 0
}

#######################################################
#(3)生成GPTN配置文件
#######################################################
function generateToml() {
        GPTNGEN=${1}
        #节点名称
        NODENAM=${2}
	#配置目录
        PCFGPAH=${3}

        if [ -f "$GPTNGEN" ]; then
                echo "Using gptn -> $GPTNGEN"
        else
            	echo "Building Configuration File"
		exit 1
        fi

        echo
        echo "##########################################################"
        echo "#####Generate configuration file using gptn tool #########"
        echo "##########################################################"
        ${GPTNGEN} dumpconfig
        if [ -f "${PCFGPAH}/ptn-config.toml" ]; then
                if [ ! -d "${PCFGPAH}/channel-artifacts/${NODENAM}" ]; then
                        mkdir -p ${PCFGPAH}/channel-artifacts/${NODENAM}
                fi

                mv -f ${PCFGPAH}/ptn-config.toml ${PCFGPAH}/channel-artifacts/${NODENAM}
        else
                echo "Generate configuration file failed,please check!"
        fi
        echo
	
	return 0
}

#######################################################
#(4)除相关的临时文件
#######################################################
function clearTmpfiles() {
	#配置目录
        PCFGPAH=${1}

        #删除临时目录log
        if [ -d "${PCFGPAH}/log" ]; then
                rm -rf ${PCFGPAH}/log
        fi
        #删除临时文件
        rm -rf ${PCFGPAH}/ptn-config.toml
        rm -rf ${PCFGPAH}/ptn-genesis.json

	return 0
}


#######################################################
#(5)修改Mediator节点中的Toml配置文件
#######################################################
function updateTomlFile() {
        #监听和合约端口
        ListenAddr=30303
        BtcHost=18332
        ContractAddress=12345

        #修改相关参数信息
        ##配置文件目录信息
        DUMPFILE=${1}
	NUM=0

        ##修改TOML配置文件中的普通参数
        n_ListenAddr="ListenAddr=\":$[${ListenAddr}+${NUM}]\""
        n_BtcHost="BtcHost=\"localhost:${BtcHost}\""
        n_ContractAddress="ContractAddress=\"127.0.0.1:${ContractAddress}\""
        n_EnableStaleProduction="EnableStaleProduction=true"

        #修改TOML配置文件中的账号参数
        account=${2}
        n_Address="Address=\"${account}\""

        #修改TOML配置文件中的密码参数
        n_Password="Password=\"1\""

        #修改TOML配置文件中的秘钥参数
        privatekey=${3}
        publickey=${4}
        n_InitPartSec="InitPartSec=\"$privatekey\""
        n_InitPartPub="InitPartPub=\"$publickey\""

        #修改配置文件
        sed -i '/^ListenAddr/c'${n_ListenAddr}'' ${DUMPFILE}
        sed -i '/^BtcHost/c'${n_BtcHost}'' ${DUMPFILE}
        sed -i '/^ContractAddress/c'${n_ContractAddress}'' ${DUMPFILE}
        sed -i '/^EnableStaleProduction/c'${n_EnableStaleProduction}'' ${DUMPFILE}
        sed -i '/^Address/c'${n_Address}'' ${DUMPFILE}
        sed -i '/^Password/c'${n_Password}'' ${DUMPFILE}
        sed -i '/^InitPartSec/c'${n_InitPartSec}'' ${DUMPFILE}
        sed -i '/^InitPartPub/c'${n_InitPartPub}'' ${DUMPFILE}

        return 0
}

#######################################################
#(6)修改创世区块文件
#######################################################
function updateGenesis() {
        #环境参数
        genesisfile=${1}

        #得到账号相关信息
        acc=`cat ${genesisfile} | jq -r '.initialMediatorCandidates[].account'`

        if [ "$acc" =  "" ]; then
                del=`cat ${genesisfile} |  jq 'del(.initialMediatorCandidates[0])'`
                add1=`echo $del | jq ".initialMediatorCandidates[.initialMediatorCandidates| length] |= . + {\"account\": \"$2\", \"initPubKey\": \"$3\", \"node\": \"$4\"}"`

                add=`echo $add1 | jq "to_entries |
                        map(if .key == \"tokenHolder\"
                                then . + {\"value\":\"$2\"}
                            else .
                            end
                        ) | from_entries"`

        else
                #echo "is not null"
                add=`cat ${genesisfile} | jq ".initialMediatorCandidates[.initialMediatorCandidates| length] |= . + {\"account\": \"$2\", \"initPubKey\": \"$3\", \"node\": \"$4\"}"`
        fi

        #删除临时文件
        if [ -f ${genesisfile} ]; then
                rm -rf ${genesisfile}
        fi
        #写入新文件
        echo $add >> temp.json
        jq -r . temp.json >> ${genesisfile}
        #
        if [ -f temp.json ]; then
                rm -rf temp.json
        fi

        return 0
}

#######################################################
#(7)生成账号信息
#######################################################
function getPtnAccount() {
	SCPATH=${1}
        Account=`${SCPATH}/scripts/getAccount.sh`
        tempinfo=`echo ${Account} | sed -n '$p'| awk '{print $NF}'`
        accountlength=35
        accounttemp=${tempinfo:0:$accountlength}
        account=`echo ${accounttemp//^M/}`
        echo ${account}
}

#######################################################
#(8)生成秘钥相关信息
#######################################################
function getPtnKeys() {
        #gptn可执行程序路径
        GPTNGEN=${1}

        key=`${GPTNGEN} mediator initdks`
        privatekeylength=44
        private=${key#*private key: }
        privatekeytemp=${private:0:$privatekeylength}
        privatekey=`echo ${privatekeytemp//^M/}`
        publickeylength=175
        public=${key#*public key: }
        publickeytemp=${public:0:$publickeylength}
        publickey=`echo ${publickeytemp//^M/}`
        echo ${privatekey}_${publickey}
}

#######################################################
#(9)生成节点相关信息
#######################################################
function getPtnNode() {
	#gptn可执行程序路径
        GPTNGEN=${1}
        b=140

        while true
        do
                info=`${GPTNGEN} nodeInfo`
                tempinfo=`echo $info | sed -n '$p'| awk '{print $NF}'`
                length=`echo ${#tempinfo}`
                nodeinfotemp=${tempinfo:0:$length}
                nodeinfo=`echo ${nodeinfotemp//^M/}`
		length=`echo ${#nodeinfo}`
                if [ "$length" -lt "$b" ]; then
                        continue
                else
                        break
                fi
        done
        echo ${nodeinfo}
}

#######################################################
#(10)更新全部初始块相关信息
#######################################################
function replaceJson() {
        #创世区块文件
        GENESIS=${1}

        #操作逻辑
        length=`cat ${GENESIS} | jq '.initialMediatorCandidates| length'`
        MinMediatorCount="MinMediatorCount"
        line=`awk "/$MinMediatorCount/{print NR}" ${GENESIS}`
        newMinMediatorCount="\"MinMediatorCount\":$length,"
        replace=`sed -e "${line}c $newMinMediatorCount" ${GENESIS}`

        rm -f ${GENESIS}
        `echo $replace >> t.json`
        jq -r . t.json >> ${GENESIS}
        rm -f t.json

        add=`cat ${GENESIS} | jq "to_entries | map(if .key == \"initialActiveMediators\"
          then . + {\"value\":$length}
          else .
          end
         ) | from_entries"`

        rm ${GENESIS}
        echo $add >> temp.json
        jq -r . temp.json >> ${GENESIS}
        rm temp.json
        echo "======modify genesis json ok======="
	return 0
}


#######################################################
#(11)
#######################################################
function addStaticNodes() {
	#创世区块文件
    	filename=${1}
	#
    	nodes=${2}
    	index=${3}
	chpah=${4}

    	content=`cat $filename`
    	acount=1
    	array="["
    	while [ $acount -le $nodes ] ;
    	do
        	if [ $acount -ne $index ];then
            		#echo $acount
            		nodeinfo=`echo $content | jq ".initialMediatorCandidates[ $[$acount-1] ].node"`
            		array="$array$nodeinfo,"
        	fi
        	let ++acount;
    	done

    	l=${#array}
    	newarr=${array:0:$[$l-1]}
    	newarr="$newarr]"
    	newStaticNodes="StaticNodes=$newarr"
    	sed -i '/^StaticNodes/c'$newStaticNodes'' ${chpah}/mediator${index}/ptn-config.toml
    	echo "=====addStaticNodes $index ok======="

	return 0
}

#######################################################
#(12)
#######################################################
function modifyStaticNodes() {
        count=0;
	#创世区块文件
        filename=${1}
	
	#循环添加pnode节点信息
        while [ ${count} -lt ${2} ] ;
        do
                echo "开始添加 ${count} 节点的pnode信息!"
                addStaticNodes ${filename} ${2} ${count} ${3}
                let ++count;
                sleep 1;
        done

    	return 0;
}

#######################################################
#(13)
#######################################################
function ExecInit() {
	##环境变量
	#channel-artifacts目录
	CHANNELPATH=${1}
	NUMNODES=${2}

        count=0
        while [ ${count} -lt ${NUMNODES} ] ;
        do
		if [ ${count} -eq 0 ]; then
			#初始化leveldb信息
			cd ${CHANNELPATH}/mediator${count}
			cp ${CHANNELPATH}/../scripts/getInit.sh .

			gptninit=`${CHANNELPATH}/../scripts/getInit.sh`
			#`echo $gptninit`
			path=`pwd`
			fullpath=${path}"/palletone/gptn/leveldb"
			if [ ! -d $fullpath ]; then
				echo "====================init err=================="
				return
			fi

			#删除临时文件和目录
			rm -rf log ${CHANNELPATH}/mediator${count}/getInit.sh
		else
			cd ${CHANNELPATH}/mediator${count}
			cp ${CHANNELPATH}/mediator0/palletone/gptn/leveldb ${CHANNELPATH}/mediator${count}/palletone/gptn -rf
		fi

		#创建日志目录
		mkdir -p ${CHANNELPATH}/mediator${count}/log

		let ++count
		sleep 1
    	done

    	echo "====================init ok====================="
    	return 0;
}
