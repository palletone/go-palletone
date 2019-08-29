##运行脚本

    ./bytn.sh 1.0.1

##说明：

    首先，该脚本默认生成有 5 个超级节点和 1 个普通全节点的本地私有链；
    然后，从 Dcoker hub 上下载 palletone/gptnnode:1.0.1 (脚本参数)镜像作为节点容器镜像，镜像标签与当前主网发布版本一致。

##修改配置文件

    需要手动修改各个节点的 ptn-config.toml 配置文件中的 VmEndpoint 的值，该值主要是作为宿主机 Docker server 的 tcp 监听端口，将 VmEndpoint 的值修改为 tcp://宿主机ip:2375，请确保宿主机 Docker server 已开启监听该端口。
    如： VmEndpoint = "tcp://192.168.152.128:2375"

##使用docker-compose 启动容器

    docker-compose up -d

##进入超级节点 1 容器

    docker exec -it mediator1 /bin/bash

##进入gptn程序控制台

    ./gptn attach

##启动节点产块

    mediator.startProduce()

##注意：

    其中，有 5 个容器名为：mediator1，mediator2，mediator3，mediator4，mediator5 为超级节点容器，1 个名为 node 是一个普通全节点。

##此时，测试网络搭建完成

##附加

    使用docker-compose 停止容器

    docker-compose stop 

    使用docker-compose 停止并移除容器

    docker-compose down



