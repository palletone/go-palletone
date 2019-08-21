#构建镜像
./start.sh 1.0.1 https://github.com/palletone/go-palletone/releases/download/v1.0.1/go-palletone_mainnet_v1.0.1_linux-amd64.tar.gz
#启动并后台运行容器
#-d:后台启动，--name 设置容器的名称
#-v /palletonedata:/palletone/data:将宿主机/palletonedate目录挂载到容器/palletone/data，方便升级时使用
docker run -d --name testnetnode -v /palletonedate:/palletone/data palletone/testnetnode:1.0.1

#进入该容器
docker exec -it testnetnode /bin/bash

#进入gptn程序console进行交互
./gptn attach palletone/testnet/gptn.ipc

#停止容器
docker stop testnetnode

#启动容器
docker start testnetnode

#进入该容器
docker exec -it testnetnode /bin/bash

#进入gptn程序console进行交互
./gptn attach palletone/testnet/gptn.ipc



#注意：没有golang相关环境
