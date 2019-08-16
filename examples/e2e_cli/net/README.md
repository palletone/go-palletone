#先下载镜像
./download-dockerimages.sh tag

#启动容器，5个节点的测试网络
docker-compose up -d

#进入容器
docker exec -it mediator1 /bin/bash

#进入gptn程序控制台
./gptn attach

#启动节点产块
mediator.startProduce()

#此时，测试网络搭建完成

#我们可以启动一个全节点
docker run -d --net=host --name normalnode palletone/normalnode:1.0.1

#进入容器
docker exec -it normalnode /bin/bash

#进入gptn程序控制台
./gptn attach
