#先下载镜像
#./bytn.sh tag
./bytn.sh 1.0.1

#启动容器，5个mediator节点和1个全节点的测试网络
docker-compose up -d

#进入容器
docker exec -it mediator1 /bin/bash

#进入gptn程序控制台
./gptn attach

#启动节点产块
mediator.startProduce()

#其中，容器名称为node是一个普通全节点

#此时，测试网络搭建完成



