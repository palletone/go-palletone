#使用脚本生成节点、下载镜像和启动5个超级节点和1个普通全节点的本地私有链
./bytn.sh 

#需要手动修改各个节点的ptn-config.toml配置文件，将 VmEndpoint 的值修改为 tcp://宿主机ip:2375，这个值是docker server监听tcp端口
如： VmEndpoint = "tcp://192.168.152.128:2375"

#使用docker-compose 启动容器
docker-compose up -d

#进入容器
docker exec -it mediator1 /bin/bash

#进入gptn程序控制台
./gptn attach

#启动节点产块
mediator.startProduce()

#其中，容器名称为node是一个普通全节点

#此时，测试网络搭建完成

#使用docker-compose 停止容器
docker-compose stop 

#使用docker-compose 停止并移除容器
docker-compose down



