#构建镜像

    ./start.sh 1.0.1 https://github.com/palletone/go-palletone/releases/download/v1.0.1/go-palletone_mainnet_v1.0.1_linux-amd64.tar.gz


#主网启动并后台运行全节点容器

    docker run -d --name mainnetgptnnode palletone/gptnnode:1.0.1

##进入该容器

    docker exec -it mainnetgptnnode /bin/bash

##进入gptn程序console进行交互

    ./gptn attach palletone/gptn.ipc

##停止容器

    docker stop mainnetgptnnode

##启动容器

    docker start mainnetgptnnode

#测试网启动并后台运行全节点容器

    docker run -d --name testnetgptnnode palletone/gptnnode:1.0.1 --testnet

#进入该容器

    docker exec -it testnetgptnnode /bin/bash

#进入gptn程序console进行交互

    ./gptn attach palletone/testnet/gptn.ipc

#主网启动并挂载文件并后台运行容器

    docker run -d --name mainnetgptnnode -v host_absolute_path:/palletone/mainnet palletone/gptnnode:1.0.1

##进入该容器

    docker exec -it mainnetgptnnode /bin/bash

##进入gptn程序console进行交互

    ./gptn attach palletone/gptn.ipc

#测试网启动并挂载文件并后台运行容器

    docker run -d --name testnetgptnnode -v host_absolute_path:/palletone/mainnet palletone/gptnnode:1.0.1 --testnet

##进入该容器

    docker exec -it testnetgptnnode /bin/bash

##进入gptn程序console进行交互

    ./gptn attach palletone/testnet/gptn.ipc
