## 使用docker-compose 启动容器

    docker-compose up -d

## 进入超级节点 1 容器

    docker exec -it mediator1 bash

## 进入gptn程序控制台
    
    cd node1
    ./gptn attach

## 启动节点产块

    mediator.startProduce()

## 注意：

    其中，有 5 个容器名为：mediator1，mediator2，mediator3，mediator4，mediator5 为超级节点容器，1 个名为 node 是一个普通全节点。

## 此时，测试网络搭建完成

## 附加
* 使用docker-compose 停止容器
    * docker-compose stop 
----
* 使用docker-compose 停止并移除容器
    * docker-compose down
    
## 本地连接私有网络，并安装用户合约
### 从 node 容器中复制获取 node7 到当前目录
* docker cp node:/go-palletone/node7 .

### 获取 mediator1 的 ip 地址
* docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' mediator1
### 获取 mediator1 的 pnode
* docker exec mediator2 cat node2/ptn-config.toml

进入目录node7，修改配置文件 ptn-config.toml，将 StaticNodes 的值修改为任意一个容器节点的 pnode 的信息

**注意:**格式要正确：["pnode://xxad x@ip:port"]
如：**["pnode://163d38776688c1d050c09aa8398ba3eb4862bc0de4b1366b7cfa41e3fbde191b593cd83f271a3dc7c40545a41df68127367c7fe6f2831bbb11d4de1a49d70df2@172.18.0.4:30305"]**

修改完成后，启动即可
* nohup ./gptn &

进入控制台
* ./gptn attach