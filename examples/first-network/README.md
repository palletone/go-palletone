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

#我们可以下载相应的gptn及创始区块等文件，启动一个全节点。
#在当前目录下，下载node.tar.gz，并解压
tar -zxvf node.tar.gz
#进入node目录，后台运行gptn
nohup ./gptn &

#进入gptn程序控制台
./gptn attach


