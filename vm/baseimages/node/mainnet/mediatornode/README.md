#构建镜像
#docker build -t palletone/mediatornode:1.0.1 .
./start.sh https://github.com/palletone/go-palletone/releases/download/v1.0.1/go-palletone_mainnet_v1.0.1_linux-amd64.tar.gz

#启动并后台运行容器
#-d:后台启动，-p 8545:8545:将宿主机8545端口映射到容器8545端口
#-v /palletonedata:/palletone/data:将宿主机/palletonedate目录挂载到容器/palletone/data，方便升级时使用
docker run -d -p 8545:8545 --name mediatornode -v /palletonedate:/palletone/data palletone/mediatornode:1.0.1 ./gptn

#进入该容器
docker exec -it mediatornode /bin/bash

#进入gptn程序console进行交互
./gptn attach palletone/gptn.ipc

#创建账户信息

#修改配置文件

#再次启动

#停止容器
docker stop mediatornode

#启动容器
docker start mediatornode

#进入该容器
docker exec -it mediatornode /bin/bash

#进入gptn程序console进行交互
./gptn attach palletone/gptn.ipc



#注意：没有golang相关环境
