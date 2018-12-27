############################################################  
#                    Pallet Cluster Building Manual        #
#                    Pallet集群搭建手册                    #
############################################################  
This document is the basic configuration document for the private network built by PalletOne. The specific steps are as follows:
本文档为PalletOne搭建私有网络的基本配置文档，具体步骤参考如下：
############################################################
#                    (1)First                              #
#                    (1)第一部分                           #
############################################################

(1)Basic Environment Configuration
(1)基础环境配置
	(1.1)Cluster planning
	(1.1)集群规划
		IP	            Name	    Role	    OS
		192.168.110.117	K8SNode01	Mediator0	Centos7
		192.168.110.118	K8SNode02	Mediator1	Centos7
		192.168.110.119	K8SNode03	Mediator2	Centos7
		192.168.110.120	K8SNode04	Mediator3	Centos7
	(1.2)Host configuration
	(1.2)主机配置
		1>Close SELinux and firewall
		1>关闭selinux和防火墙
		#sudo sed -i 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/selinux/config
		#sudo systemctl stop firewalld.service && systemctl disable firewalld.service
		#
		2>Configure Host Time, Time Zone, System Language
		2>配置主机时间、时区、系统语言
		##Modify time zone
		##修改时区
		#ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
		#
		##Modifying the System Language Environment
		##修改系统语言环境
		#sudo echo 'LANG="en_US.UTF-8"' >> /etc/profile;source /etc/profile
		#
		##Configure host NTP time synchronization
		##配置主机NTP时间同步
		#ntpdate -u cn.pool.ntp.org
		#
		3>performance tuning of Kernel
		3>Kernel性能调优
		#cat >> /etc/sysctl.conf<<EOF
		net.ipv4.ip_forward=1
		net.ipv4.neigh.default.gc_thresh1=4096
		net.ipv4.neigh.default.gc_thresh2=6144
		net.ipv4.neigh.default.gc_thresh3=8192
		EOF
		#sysctl -p
		#
		4>Add host information
		4>添加主机信息到/etc/hosts
		#sudo hostnamectl set-hostname k8snode01     #Each node is modified with a different name[各个节点进行修改，名字不同]
		#vi /etc/hosts
		192.168.110.117 k8snode01
		192.168.110.118 k8snode02
		192.168.110.119 k8snode03
		192.168.110.120 k8snode04
	(1.3)Configure Docker
	(1.3)配置Docker
		0>Installation Pre-Dependency
		0>安装前置依赖
		#yum -y install expect spawn jq
		#
		1>Adding users (optional)
		1>添加用户(可选)
		#sudo adduser k8sdocker
		#
		2>Setting passwords for new users(optional)
		2>为新用户设置密码(可选)
		#sudo passwd k8sdocker
		#
		3>Setting passwords for new users(optional)
		3>为新用户添加sudo权限(可选)
		#sudo echo 'k8sdocker ALL=(ALL) ALL' >> /etc/sudoers
		#
		4>Uninstall old version of Docker software
		4>卸载旧版本Docker软件
		#sudo yum remove docker \
		              docker-client \
		              docker-client-latest \
		              docker-common \
		              docker-latest \
		              docker-latest-logrotate \
		              docker-logrotate \
		              docker-selinux \
		              docker-engine-selinux \
		              docker-engine \
		              container*
		#
		5>Define installation version
		5>定义安装版本
			5.1># step 1: install the necessary system tools
			5.1># step 1: 安装必要的一些系统工具
			#sudo yum update -y
			#sudo yum install -y yum-utils device-mapper-persistent-data lvm2 bash-completion
			#export docker_version=17.03.2
			#
			5.2># Step 2: Adding Software Source Information
			5.2># Step 2: 添加软件源信息
			#sudo yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
			#
			5.3># Step 3: Update and install Docker-CE
			5.3># Step 3: 更新并安装 Docker-CE
			#sudo yum makecache all
			#version=$(yum list docker-ce.x86_64 --showduplicates | sort -r|grep ${docker_version}|awk '{print $2}')
			#sudo yum -y install --setopt=obsoletes=0 docker-ce-${version} docker-ce-selinux-${version}
			#
			5.4># Step 4: If you have installed a high-level version of Docker, you can downgrade the installation (optional)
			5.4># Step 4: 如果已经安装高版本Docker,可进行降级安装(可选)
			#yum downgrade --setopt=obsoletes=0 -y docker-ce-${version} docker-ce-selinux-${version}
			#
			5.5># Step 5: Set up boot start
			5.5># Step 5: 设置开机启动
			#sudo systemctl enable docker && systemctl start docker.service
			#
		6>Docker configuration
		6>Docker配置
		##Configure image acceleration address
		##配置镜像加速地址
		#vi /etc/docker/daemon.json
		{
		  "registry-mirrors": ["https://kri93zmv.mirror.aliyuncs.com"]
		}
		备注：如果出现：Failed to start Docker Application Container Engine，则执行如下语句：
		#rm /etc/docker/key.json 
		#rm -rf /var/lib/docker/
	(1.4)Configure Palletone
	(1.4)配置Palletone
		0>Configure GOPATH environment variables
		0>配置GOPATH环境变量
		##Install the specified go1.10.4. version
		##安装指定go1.10.4.版本
		#go version
		#
		##Configuring environment variables
		##配置环境变量
		#vi /etc/profile
		export GOROOT=/opt/go
		export GOPATH=/opt/gopath
		export PATH=$PATH:/opt/go/bin:/opt/gopath/bin
		#source /etc/profile
		#
		1>Download the code and compile it
		1>下载代码，并对其进行编译
		#cd $GOPATH
		#mkdir -p src/github.com/palletone
		#cd $GOPATH/src/github.com/palletone
		#go get -u github.com/palletone/adaptor
		#go get -u github.com/palletone/btc-adaptor
		#go get -u github.com/palletone/eth-adaptor
		#git clone https://github.com/palletone/go-palletone.git
		#ll
			drwxr-xr-x  3 root root  138 Dec 24 02:14 adaptor
			drwxr-xr-x  4 root root  230 Dec 24 02:15 btc-adaptor
			drwxr-xr-x  5 root root 4096 Dec 24 02:15 eth-adaptor
			drwxr-xr-x 25 root root 4096 Dec 24 20:39 go-palletone
		#
		#cd go-palletone/
		#make all
		#
		##Note: Be sure to remember the path of the gptn executable
		##备注：务必记住gptn可执行程序的路径($PROJECT/build/bin/gptn)
		#
		2>Give privileges to scripts
		2>为脚本赋予权限
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli
		#chmod -R 775 *
		#
		##Modify the location of the gptn executable file in getAccount.sh in the scripts directory to the absolute path under build/bin
		##将scripts目录中的getAccount.sh中的gptn可执行文件位置进行修改为build/bin下面的绝对路径
									#!/usr/bin/expect
									#!/bin/bash
									#产生账号信息
									set timeout 30
									#需要配置gptn的绝对路径地址
									#spawn "gptn's dir" account new
									spawn /opt/gopath/src/github.com/palletone/go-palletone/build/bin/gptn account new
									expect "Passphrase:"
									send "1\r"
									expect "Repeat passphrase:"
									send "1\r"
									interact
		#
		##Modify the location of the gptn executable file in getInit.sh in the scripts directory to the absolute path under mediator0
		##将scripts目录中的getInit.sh中的gptn可执行文件位置进行修改为mediator0下面的绝对路径
									#!/usr/bin/expect
									set timeout 30
									#输入gptn在Mediator0的绝对路径
									spawn /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator0/gptn init
									expect "Passphrase:"
									send "1\r"
									interact
		#
		3>Generating configuration path information for N nodes
		3>生成N个节点的配置路径信息
		#./generateArtifacts.sh 3               #The latter parameter can only be odd[后面的参数只能为奇数]
		#ls -l
		##Note: After executing the script, the channel-artifacts directory is generated. If the execution fails, the directory is deleted.
		##备注：执行该脚本后生成channel-artifacts目录，若执行失败，则将该目录进行删除
############################################################
#                    (2)Second                             #
#                    (2)第二部分                           #
############################################################

(2)Configuring Super Nodes
(2)配置超级节点
	(2.1)Generate Docker image [optional]
	(2.1)生成Docker镜像【可选】
		1>Generating docker image related files
		1>生成docker镜像相关文件
		#cd $GOPATH/src/github.com/go-palletone
		#make docker
		#cd build/images/gptn
		#
		2>Generate an image and upload it to dockerhub
		2>生成镜像，并上传到dockerhub
		#docker build -t palletone/pallet-gptn:0.6 .
		#docker images
		#
		3>Log in to dockerhub
		3>登录dockerhub
		#docker login
		#docker push palletone/pallet-gptn:0.6
		#
		##Query the image in dockerhub
		##在dockerhub查询该镜像
		#
	(2.2)配置Mediator0
	(2.2)Configure Mediator0
	  ##Enter the Mediator0 directory and configure accordingly. The specific configuration information is as follows:
		##进入Mediator0目录进行相应的配置，具体配置信息如下所示：
		1>Enter mediator0 directory
		1>进入mediator0目录
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli
		#cd channel-artifacts/mediator0
		#ls -l
					-rwxr-xr-x 1 root root      291 Dec 25 01:25 getInit.sh
					-rwxr-xr-x 1 root root 49189104 Dec 25 01:25 gptn
					drwxr-xr-x 2 root root       38 Dec 25 01:25 log
					drwx------ 5 root root       68 Dec 25 02:41 palletone
					-rw-r--r-- 1 root root     3343 Dec 25 01:26 ptn-config.toml
					-rw-r--r-- 1 root root     2854 Dec 25 01:26 ptn-genesis.json
		2>Modify toml file
		2>修改toml文件
		#vi ptn-config.toml
			[Node]
			DataDir = "/var/palletone/production"
			[Log]
			OutputPaths = ["stdout", "/var/palletone/log/all.log"]
			ErrorOutputPaths = ["stderr", "/var/palletone/log/error.log"]
			[Contract]
			ContractFileSystemPath = "/var/palletone/production/chaincodes"
			[P2P]
			MaxPeers = 25
			NoDiscovery = false
			BootstrapNodes = []
			StaticNodes=["pnode://3ea34ff09489627399bbeac8d3af93b34981afc623228210bd49c8ce11860f78c736aa3721ebb91aec76353a3b93ee6a2aadd05337ab0723a71a7c9f68947144@mediator0:30303","pnode://01f20de81a80738b30d944a756ade9f4222f95a696d45b451aed596eefa204f3c8ae98305363feceeb28f5c140a6736118f59c81716c0cdd123365cad8a528eb@mediator1:30303","pnode://2a891ee523a40961c0760871be0613551aab45ad7a4ecd23369b713601228173b6e91d4ce748a2cbb571ae0c9b4d47ce605500ad3785cbadeb9ca8ba1a412f6e@mediator2:30303"]
		##Note: Change the relative path in the configuration file to absolute path, replace the IP address of gptn in StaticNodes in p2p, and change it to [container name] if it runs locally.
		##备注:将配置文件中的相对路径修改为绝对路径；将p2p中的StaticNodes中替换为gptn所在的IP地址；若为本机运行，将其修改为[容器名称];
		3>Modify JSON file
		3>修改json文件
		#vi ptn-genesis.json
		#
		##Note: No modification is required for the time being.
		##备注：暂时不用修改
	(2.3)配置Mediator1
	(2.3)Configure Mediator1
		##Enter the Mediator1 directory and configure it accordingly. The specific configuration information is as follows:
		##进入Mediator1目录进行相应的配置，具体配置信息如下所示：
		1>Enter mediator1 directory
		1>进入mediator1目录
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli
		#cd channel-artifacts/mediator1
		#ls -l
					drwxr-xr-x 2 root root       38 Dec 25 01:25 log
					drwx------ 5 root root       68 Dec 25 02:41 palletone
					-rw-r--r-- 1 root root     3343 Dec 25 01:26 ptn-config.toml
		2>Modify toml file
		2>修改toml文件
		#vi ptn-config.toml
			[Node]
			DataDir = "/var/palletone/production"
			[Log]
			OutputPaths = ["stdout", "/var/palletone/log/all.log"]
			ErrorOutputPaths = ["stderr", "/var/palletone/log/error.log"]
			[Contract]
			ContractFileSystemPath = "/var/palletone/production/chaincodes"
			[P2P]
			MaxPeers = 25
			NoDiscovery = false
			BootstrapNodes = []
			StaticNodes=["pnode://3ea34ff09489627399bbeac8d3af93b34981afc623228210bd49c8ce11860f78c736aa3721ebb91aec76353a3b93ee6a2aadd05337ab0723a71a7c9f68947144@mediator0:30303","pnode://2a891ee523a40961c0760871be0613551aab45ad7a4ecd23369b713601228173b6e91d4ce748a2cbb571ae0c9b4d47ce605500ad3785cbadeb9ca8ba1a412f6e@mediator2:30303"]
		##Note: 
				Update pnode information to Mediator0 and mediator2;
				Modify the relative path in the configuration file to absolute path；
				Replace the IP address of gptn in StaticNodes in p2p, and change it to [container name] if it runs locally.
		##备注:
				将pnode信息更新为mediator0和mediator2；
				将配置文件中的相对路径修改为绝对路径；
				将p2p中的StaticNodes中替换为gptn所在的IP地址；若为本机运行，将其修改为[容器名称];
	(2.4)配置Mediator2
	(2.4)Configure Mediator2
		##Enter the Mediator2 directory and configure it accordingly. The specific configuration information is as follows:
		##进入Mediator2目录进行相应的配置，具体配置信息如下所示：
		1>Enter mediator2 directory
		1>进入mediator2目录
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli
		#cd channel-artifacts/mediator2
		#ls -l
					drwxr-xr-x 2 root root       38 Dec 25 01:25 log
					drwx------ 5 root root       68 Dec 25 02:41 palletone
					-rw-r--r-- 1 root root     3343 Dec 25 01:26 ptn-config.toml
		2>Modify toml file
		2>修改toml文件
		#vi ptn-config.toml
			[Node]
			DataDir = "/var/palletone/production"
			[Log]
			OutputPaths = ["stdout", "/var/palletone/log/all.log"]
			ErrorOutputPaths = ["stderr", "/var/palletone/log/error.log"]
			[Contract]
			ContractFileSystemPath = "/var/palletone/production/chaincodes"
			[P2P]
			MaxPeers = 25
			NoDiscovery = false
			BootstrapNodes = []
			StaticNodes=["pnode://3ea34ff09489627399bbeac8d3af93b34981afc623228210bd49c8ce11860f78c736aa3721ebb91aec76353a3b93ee6a2aadd05337ab0723a71a7c9f68947144@mediator0:30303","pnode://2a891ee523a40961c0760871be0613551aab45ad7a4ecd23369b713601228173b6e91d4ce748a2cbb571ae0c9b4d47ce605500ad3785cbadeb9ca8ba1a412f6e@mediator1:30303"]
		##Note: 
				Update pnode information to Mediator0 and mediator1;
				Modify the relative path in the configuration file to absolute path；
				Replace the IP address of gptn in StaticNodes in p2p, and change it to [container name] if it runs locally.
		##备注:
				将pnode信息更新为mediator0和mediator2；
				将配置文件中的相对路径修改为绝对路径；
				将p2p中的StaticNodes中替换为gptn所在的IP地址；若为本机运行，将其修改为[容器名称];
	(2.5)Configure docker template
	(2.5)配置docker模板
		##Enter the e2e_cli directory and configure it accordingly. The specific configuration information is as follows:
		##进入e2e_cli目录进行相应的配置，具体配置信息如下所示：
		1>Enter e2e_cli directory
		1>进入e2e_cli目录
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli
		#
		2>Configuration template file
		2>配置模板文件
		#vi docker-compose-e2e.yaml
		#
		##Note: Refer to the file under e2e_cli directory.
		##备注:参考e2e_cli目录下面的该文件；
		3>Start container
		3>启动容器
		##Startup container in the background
		##后台启动容器
		#docker-compose -f docker-compose-e2e.yaml up -d
		#
		##Query running container
		##查询运行的容器
		#docker ps
		#
		4>Stop running the container (optional)
		4>停止运行容器(可选)
		#docker-compose -f docker-compose-e2e.yaml down
		#
############################################################
#                    (3)Third                              #
#                    (3)第三部分                           #
############################################################

(3)Configuration data node
(3)配置数据节点
	(3.1)Configuration data node
	(3.1)配置数据节点
		1>Copy mediator0 to mediator3
		1>拷贝mediator0为mediator3
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts
		#
		#cp -rf mediator0 mediator3
		#
		2>Modify the toml configuration file
		2>修改toml配置文件
		#cd mediator3
		#vi ptn-config.toml
			[MediatorPlugin]
			EnableStaleProduction=false
		3>Add mediator3 configuration information to the template file in the e2e_cli directory.
		3>在e2e_cli目录的模板文件中添加mediator3配置信息
		#vi docker-compose-e2e.yaml
		#
		4>Start the service based on the template file
		4>根据模板文件启动服务
		#docker-compose -f docker-compose-e2e.yaml up -d
		#docker ps
		#
		5>Enter the Mediator 3 directory to view the logs
		5>进入mediator3目录查看日志
		#cd mediator3/log
		#tail -f all.log
############################################################
#                    (4)Four                               #
#                    (4)第四部分                           #
############################################################

(4)End
(4)结尾
