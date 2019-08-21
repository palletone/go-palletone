#使用脚本生成节点、下载镜像和启动5个超级节点和1个普通全节点的本地私有链
#./bytn.sh
./bytn.sh 

#进入容器
docker exec -it mediator1 /bin/bash

#进入gptn程序控制台
./gptn attach

#启动节点产块
mediator.startProduce()

#其中，容器名称为node是一个普通全节点

#此时，测试网络搭建完成



