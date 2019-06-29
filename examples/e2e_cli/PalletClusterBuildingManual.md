
#                    Pallet Cluster Building Manual        #
This document is the basic configuration document for the private network built by PalletOne. The specific steps are as follows:

## (1) Basic Environment Configuration
The basic configuration commands in CentOS are as follows:
### 	(1.1) Cluster planning
		IP	        Name	    	Role	    	OS
		192.168.110.117	K8SNode01	Mediator0	Centos7
		192.168.110.118	K8SNode02	Mediator1	Centos7
		192.168.110.119	K8SNode03	Mediator2	Centos7
		192.168.110.120	K8SNode04	Client01	Centos7
### 	(1.2) Host configuration
#### 		(1.2.1) Turn off selinux and firewall
		#sudo sed -i 's/SELINUX=enforcing/SELINUX=disabled/g' /etc/selinux/config
		#sudo systemctl stop firewalld.service && systemctl disable f irewalld.service
		#
#### 		(1.2.2) Configure host time, time zone, system language
		##Modify time zone
		#ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
		#
		##Modifying the System Language Environment
		#sudo echo 'LANG="en_US.UTF-8"' >> /etc/profile;source /etc/profile
		#
		##Configure host NTP time synchronization
		#ntpdate -u cn.pool.ntp.org
		#
#### 		(1.2.3) Kernel performance tuning
		#cat >> /etc/sysctl.conf<<EOF
		net.ipv4.ip_forward=1
		net.ipv4.neigh.default.gc_thresh1=4096
		net.ipv4.neigh.default.gc_thresh2=6144
		net.ipv4.neigh.default.gc_thresh3=8192
		EOF
		#sysctl -p
		#
#### 		(1.2.4) Add host information
		#sudo hostnamectl set-hostname k8snode01     #Each node is modified with a different name
		#vi /etc/hosts
		192.168.110.117 k8snode01
		192.168.110.118 k8snode02
		192.168.110.119 k8snode03
		192.168.110.120 k8snode04
### 	(1.3) Configuring Docker
####		(1.3.0) Installation Pre-Dependency
		#yum -y install expect spawn jq docker-compose
		#
#### 		(1.3.1) Add user (optional)
		#sudo adduser k8sdocker
		#
#### 		(1.3.2) Set a password for new users (optional)
		#sudo passwd k8sdocker
		#
#### 		(1.3.3) Add sudo privileges for new users (optional)
		#sudo echo 'k8sdocker ALL=(ALL) ALL' >> /etc/sudoers
		#
#### 		(1.3.4) Uninstall old Docker software
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
#### 		(1.3.5) Define the installation version
		##Step 1: Install some necessary system tools
		#sudo yum update -y
		#sudo yum install -y yum-utils device-mapper-persistent-data lvm2 bash-completion
		#export docker_version=17.03.2
		#
		##Step 2: Adding Software Source Information
		#sudo yum-config-manager --add-repo http://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
		#
		##Step 3: Update and install Docker-CE
		#sudo yum makecache all
		#version=$(yum list docker-ce.x86_64 --showduplicates | sort -r|grep ${docker_version}|awk '{print $2}')
		#sudo yum -y install --setopt=obsoletes=0 docker-ce-${version} docker-ce-selinux-${version}
		#
		##Step 4: If you have already installed a high version of Docker, you can perform a downgrade installation (optional)
		#yum downgrade --setopt=obsoletes=0 -y docker-ce-${version} docker-ce-selinux-${version}
		#
		##Step 5: Set up boot start
		#sudo systemctl enable docker && systemctl start docker.service
		#
#### 		(1.3.6) Docker configuration
		##Configure image acceleration address
		#vi /etc/docker/daemon.json
		{
		  "registry-mirrors": ["https://kri93zmv.mirror.aliyuncs.com"]
		}
		Note: If it appears:Failed to start Docker Application Container Engine，then execute the following statement:
		#rm /etc/docker/key.json 
		#rm -rf /var/lib/docker/
### 	(1.4) Configure Palletone
####		(1.4.0) Configuring the GOPATH environment variable
		##Install the specified go1.10.4 version
		#go version
			go version go1.10.4 linux/amd64
		#
		##Configuring environment variables
		#vi /etc/profile
		export GOROOT=/opt/go
		export GOPATH=/opt/gopath
		export PATH=$PATH:/opt/go/bin:/opt/gopath/bin
		#source /etc/profile
		#
#### 		(1.4.1) Download the code and compile it
		#cd $GOPATH
		#mkdir -p src/github.com/palletone
		#cd $GOPATH/src/github.com/palletone
		#go get -u github.com/palletone/adaptor
		#go get -u github.com/palletone/btc-adaptor
		#go get -u github.com/palletone/eth-adaptor
		#git clone https://github.com/palletone/go-palletone.git
		#ls -l
			drwxr-xr-x  3 root root  138 Dec 24 02:14 adaptor
			drwxr-xr-x  4 root root  230 Dec 24 02:15 btc-adaptor
			drwxr-xr-x  5 root root 4096 Dec 24 02:15 eth-adaptor
			drwxr-xr-x 25 root root 4096 Dec 24 20:39 go-palletone
		#
		#cd go-palletone/
		#make all
		#
		##Note: Be sure to remember the path of the gptn executable
		#
#### 		(1.4.2) Give permissions to scripts
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli
		#chmod -R 775 *
		#
		##Modify the gptn executable location in getAccount.sh in the scripts directory to the absolute path under build/bin
			#!/usr/bin/expect
			#!/bin/bash
			#产生账号信息
			set timeout 30
			#需要配置gptn的绝对路径地址
			#spawn "gptn's dir" account new
			spawn /opt/gopath/src/github.com/palletone/go-palletone/build/bin/gptn account new
			expect "Passphrase:"
			send "1\n"
			expect "Repeat passphrase:"
			send "1\n"
			interact
		#
		##Modify the gptn executable location in getInit.sh in the scripts directory to the absolute path below mediator0
			#!/usr/bin/expect
			set timeout 30
			#输入gptn在Mediator0的绝对路径
			spawn /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator0/gptn init
			expect "Passphrase:"
			send "1\n"
			interact
		#
#### 		(1.4.3) Generating configuration path information for N nodes
		#./generateArtifacts.sh 3               #The following parameters can only be odd
		#ls -l
		##Note: The channel-artifacts directory is generated after executing the script. If the execution fails, the directory is deleted.

## (2) Configuring Super Nodes
### 	(2.1) Generate Docker image [optional]
#### 		(2.1.1) Generate docker image related files
		#cd $GOPATH/src/github.com/go-palletone
		#make docker
		#cd build/images/gptn
		#
#### 		(2.1.2) Generate an image and upload it to dockerhub
		#docker build -t palletone/pallet-gptn:0.6 .      #Generate version 0.6
		#docker build -t palletone/pallet-gptn     .      #Generate version latest
		#docker images
			REPOSITORY                   TAG                 IMAGE ID            CREATED             SIZE
			palletone/pallet-gptn        0.6                 3ff3b7de24c9        2 days ago          406MB
			palletone/pallet-gptn        latest              3ff3b7de24c9        2 days ago          406MB
		#
#### 		(2.1.3) Log in to dockerhub
		#docker login
		#docker push palletone/pallet-gptn:0.6
		#docker push palletone/pallet-gptn
		##Query the image in dockerhub
		#
### 	(2.2) Configuring Mediator0
	Enter the Mediator0 directory and configure accordingly. The specific configuration information is as follows:
#### 		(2.2.1) Enter the mediator0 directory
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli
		#cd channel-artifacts/mediator0
		#ls -l
			-rwxr-xr-x 1 root root      291 Dec 25 01:25 getInit.sh
			-rwxr-xr-x 1 root root 49189104 Dec 25 01:25 gptn
			drwxr-xr-x 2 root root       38 Dec 25 01:25 log
			drwx------ 5 root root       68 Dec 25 02:41 palletone
			-rw-r--r-- 1 root root     3343 Dec 25 01:26 ptn-config.toml
			-rw-r--r-- 1 root root     2854 Dec 25 01:26 ptn-genesis.json
#####		(2.2.2) Initialize the database		
		#./gptn init
			Passphrase:                   #Enter a password, such as entering a number 1;
		#
		##Copy the database information of mediator0 to other node directories
		#cd ..
		#cp -rf mediator0/palletone/gptn/leveldb mediator1/palletone/gptn/leveldb
		#cp -rf mediator0/palletone/gptn/leveldb mediator2/palletone/gptn/leveldb
		#cd mediator0
##### 		(2.2.3) Modify the toml file
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
		##Note: Change the relative path in the configuration file to an absolute path; Replace the StaticNodes in p2p with the IP address where gptn is located; If it is running on this machine, modify it to the container name.[备注:将配置文件中的相对路径修改为绝对路径；将p2p中的StaticNodes中替换为gptn所在的IP地址；若为本机运行，将其修改为容器名称]
		##;
##### 		(2.2.4) Modify the json file
		#vi ptn-genesis.json
		#
		##Remarks: No need to modify temporarily
### 	(2.3)Configuring Mediator1
	Enter the Mediator1 directory and configure it accordingly. The specific configuration information is as follows:
#### 		(2.3.1) Enter the mediator1 directory
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli
		#cd channel-artifacts/mediator1
		#ls -l
			drwxr-xr-x 2 root root       38 Dec 25 01:25 log
			drwx------ 5 root root       68 Dec 25 02:41 palletone
			-rw-r--r-- 1 root root     3343 Dec 25 01:26 ptn-config.toml
#### 		(2.3.2) Modify the toml file
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
		##Note: Update the pnode information to mediator0 and mediator2; Change the relative path in the configuration file to an absolute path; Replace the StaticNodes in p2p with the IP address where gptn is located; If it is running on the machine, modify it to the container name.[备注:将pnode信息更新为mediator0和mediator2；将配置文件中的相对路径修改为绝对路径；将p2p中的StaticNodes中替换为gptn所在的IP地址；若为本机运行，将其修改为容器名称]
								
### 	(2.4) Configuring Mediator2
	Enter the Mediator2 directory and configure it accordingly. The specific configuration information is as follows:
#### 		(2.4.1) Enter the mediator2 directory
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli
		#cd channel-artifacts/mediator2
		#ls -l
			drwxr-xr-x 2 root root       38 Dec 25 01:25 log
			drwx------ 5 root root       68 Dec 25 02:41 palletone
			-rw-r--r-- 1 root root     3343 Dec 25 01:26 ptn-config.toml
#### 		(2.4.2) Modify the toml file
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
		##Remarks: Update the pnode information to mediator0 and mediator2; Change the relative path in the configuration file to an absolute path; Replace the StaticNodes in p2p with the IP address where gptn is located; If it is running on the machine, modify it to the container name.[备注:将pnode信息更新为mediator0和mediator2；将配置文件中的相对路径修改为绝对路径；将p2p中的StaticNodes中替换为gptn所在的IP地址；若为本机运行，将其修改为容器名称]		
				
### 	(2.5) Configuring the docker template
	Go to the e2e_cli directory to perform the corresponding configuration. The detailed configuration information is as follows:
#### 		(2.5.1) Go to the e2e_cli directory
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli
		#
#### 		(2.5.2) Configuration template file
		#vi docker-compose-e2e.yaml
		#
		##Note: Refer to the file under e2e_cli directory.[备注:参考e2e_cli目录下面的该文件；]
#### 		(2.5.3) Start container
		##Startup container in the background
		#docker-compose -f docker-compose-e2e.yaml up -d
		#
		##Query running container
		#docker ps
		#
#### 		(2.5.4) Stop running the container (optional)
		#docker-compose -f docker-compose-e2e.yaml down
		#

##  (3) Configuration data node
### 	(3.1) Configuration data node
#### 		(3.1.1) Copy mediator0 to client01
		#cd $GOPATH/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts
		#
		#cp -rf mediator0 client01
		#
#### 		(3.1.2) Modify the toml configuration file
		#cd client01
		#vi ptn-config.toml
			[MediatorPlugin]
			EnableStaleProduction=false
#### 		(3.1.3) Add client01 configuration information to the template file in the e2e_cli directory.
		#vi docker-compose-e2e.yaml
		#
#### 		(3.1.4) Start the service based on the template file
		#docker-compose -f docker-compose-e2e.yaml up -d
		#docker ps
		#
#### 		(3.1.5) Enter the client01 directory to view the logs
		#cd client01/log
		#tail -f all.log

## (4) End
After the above configuration, the content of the docker-compose-e2e.yaml file is:
	
	version: '2'
	services:
	  mediator0:
	    container_name: mediator0
	    image: palletone/pallet-gptn
	    working_dir: /opt/gopath/src/github.com/palletone/go-palletone
	    volumes:
	     - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator0/ptn-genesis.json:/var/palletone/conf/ptn-genesis.json
	     - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator0/ptn-config.toml:/var/palletone/conf/ptn-config.toml
	     - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator0/palletone:/var/palletone/production
	     - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator0/log:/var/palletone/log
	    command: /var/palletone/conf/ptn-config.toml
	    ports:
	     - 8545:8545
	     - 8546:8546
	     - 8080:8080
	     - 30303:30303
	     - 18332:18332
	     - 12345:12345

	  mediator1:
	    container_name: mediator1
	    image: palletone/pallet-gptn
	    working_dir: /opt/gopath/src/github.com/palletone/go-palletone
	    volumes:
	     - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator1/ptn-genesis.json:/var/palletone/conf/ptn-genesis.json
	     - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator1/ptn-config.toml:/var/palletone/conf/ptn-config.toml
	     - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator1/palletone:/var/palletone/production
	     - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator1/log:/var/palletone/log
	    command: /var/palletone/conf/ptn-config.toml
	    ports:
	     - 8555:8545
	     - 8556:8546
	     - 8081:8080
	     - 30304:30303
	     - 18342:18332
	     - 12355:12345

	  mediator2:
	   container_name: mediator2
	   image: palletone/pallet-gptn
	   working_dir: /opt/gopath/src/github.com/palletone/go-palletone
	   volumes:
	    - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator2/ptn-genesis.json:/var/palletone/conf/ptn-genesis.json
	    - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator2/ptn-config.toml:/var/palletone/conf/ptn-config.toml
	    - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator2/palletone:/var/palletone/production
	    - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/mediator2/log:/var/palletone/log
	   command: /var/palletone/conf/ptn-config.toml
	   ports:
	    - 8565:8545
	    - 8566:8546
	    - 8082:8080
	    - 30305:30303
	    - 18352:18332
	    - 12365:12345

	  client01:
	   container_name: client01
	   image: palletone/pallet-gptn
	   working_dir: /opt/gopath/src/github.com/palletone/go-palletone
	   volumes:
	    - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/client01/ptn-genesis.json:/var/palletone/conf/ptn-genesis.json
	    - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/client01/ptn-config.toml:/var/palletone/conf/ptn-config.toml
	    - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/client01/palletone:/var/palletone/production
	    - /opt/gopath/src/github.com/palletone/go-palletone/examples/e2e_cli/channel-artifacts/client01/log:/var/palletone/log
	   command: /var/palletone/conf/ptn-config.toml
	   ports:
	    - 8575:8545
	    - 8576:8546
	    - 8083:8080
	    - 30306:30303
	    - 18362:18332
	    - 12375:12345
