## 运行明显生成多节点
    ./start.sh 5
    5 为构建的 mediator 节点数量

## 说明：

    首先，该脚本默认生成有 5 个超级节点和 2 个普通全节点的本地私有链，一个普通全节点使用容器自动运行，另一个在本地需要手动运行。

## 使用docker-compose 启动容器

    docker-compose up -d

## 进入超级节点 1 容器

    docker exec -it mediator1 bash

## 进入gptn程序控制台

    gptn attach

## 启动节点产块

    mediator.startProduce()

## 注意：

    以 mediator 节点数量 = 5 为例，其中，有 5 个容器名为：mediator1，mediator2，mediator3，mediator4，mediator5 为超级节点容器，1 个名为 node 是一个普通全节点。

## 此时，测试网络搭建完成

## 附加
* 使用docker-compose 停止容器
    * docker-compose stop 
----
* 使用docker-compose 停止并移除容器
    * docker-compose down
    
## 本地 node7 连接私有网络，即可进行用户合约的安装、部署、调用和停止
进入目录node7，修改配置文件 ptn-config.toml，将 StaticNodes 的值修改为任意一个容器节点的 pnode 的信息  
**注意:**格式要正确：["pnode://xxx@ip:port"]
如：**["pnode://163d38776688c1d050c09aa8398ba3eb4862bc0de4b1366b7cfa41e3fbde191b593cd83f271a3dc7c40545a41df68127367c7fe6f2831bbb11d4de1a49d70df2@172.18.0.4:30305"]**  
修改完成后，启动即可
* nohup ./gptn &

进入控制台
* ./gptn attach