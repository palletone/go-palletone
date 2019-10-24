# Supported tags and respective Dockerfile links

- [latest](https://github.com/palletone/go-palletone/blob/master/images/node/Dockerfile)
- [1.0.3](https://github.com/palletone/go-palletone/blob/master/images/node/Dockerfile)
- [1.0.2](https://github.com/palletone/go-palletone/blob/master/images/node/Dockerfile)
- [1.0.1](https://github.com/palletone/go-palletone/blob/master/images/node/Dockerfile)
- [1.0.0](https://github.com/palletone/go-palletone/blob/master/images/node/Dockerfile)

# What is gptn

gptn is used as the base image for the go-palletone node,such as the main network,test network,and the local private network...    

# How to use this image

- ### 作为主网节点

- ### 普通全节点、超级节点或者陪审员节点且不需要挂载文件

  - 启动容器：docker run -d --network gptn-net --name mainnetgptn palletone/gptn:1.0.3

  - 进入容器：docker exec -it mainnetgptn bash

  - 再进入 gptn 控制台：./gptn attach

    **注意：**

    - 若提示：docker: Error response from daemon: network gptn-net not found，需要创建该网络，使用命令：**docker network create gptn-net**
    - 如果是作为超级节点或者是陪审员节点，使用以下命令开启容器
      - docker run -d --network gptn-net --name mainnetgptn -v /var/run/docker.sock:/var/run/docker.sock palletone/gptn:1.0.3

- ### 普通全节点、超级节点或者陪审员节点且需要挂载文件

  - 启动容器：docker run -d --network gptn-net --name mainnetgptn -v host_absolute_path/palletone:/go-palletone/palletone -v host_absolute_path/ptn-config.toml:/go-palletone/ptn-config.toml palletone/gptn:1.0.3

  - 进入容器：docker exec -it mainnetgptn bash

  - 再进入gptn控制台：./gptn attach    

    **注意：**

    - 若提示：docker: Error response from daemon: network gptn-net not found，需要创建该网络，使用命令：**docker network create gptn-net**
    - 如果是作为超级节点或者是陪审员节点，使用以下命令开启容器
      - docker run -d --network gptn-net --name mainnetgptn -v host_absolute_path/palletone:/go-palletone/palletone -v host_absolute_path/ptn-config.toml:/go-palletone/ptn-config.toml -v /var/run/docker.sock:/var/run/docker.sock palletone/gptn:1.0.3

  ------

- ## 作为测试网节点

- ### 普通全节点、超级节点或者陪审员节点且不需要挂载文件

  - 启动容器：docker run -d --network gptn-net --name testnetgptn palletone/gptn:1.0.3 --testnet

  - 进入容器：docker exec -it testnetgptn bash

  - 再进入gptn控制台：./gptn attach palletone/testnet/gptn.ipc

    **注意：**

    - 若提示：docker: Error response from daemon: network gptn-net not found，需要创建该网络，使用命令：**docker network create gptn-net**
    - 如果是作为超级节点或者是陪审员节点，使用以下命令开启容器
      - docker run -d --network gptn-net --name testnetgptn -v /var/run/docker.sock:/var/run/docker.sock palletone/gptn:1.0.3 --testnet

- ### 普通全节点、超级节点或者陪审员节点且需要挂载文件

  - 启动容器：docker run -d --network gptn-net --name testnetgptn  -v host_absolute_path/palletone:/go-palletone/palletone -v host_absolute_path/ptn-config.toml:/go-palletone/ptn-config.toml palletone/gptn:1.0.3 --testnet

  - 进入容器：docker exec -it testnetgptn bash

  - 再进入gptn控制台：./gptn attach palletone/testnet/gptn.ipc

    **注意：**

    - 若提示：docker: Error response from daemon: network gptn-net not found，需要创建该网络，使用命令：**docker network create gptn-net**
    - 如果是作为超级节点或者是陪审员节点，使用以下命令开启容器
      - docker run -d --network gptn-net --name testnetgptn -v host_absolute_path/palletone:/go-palletone/palletone -v host_absolute_path/ptn-config.toml:/go-palletone/ptn-config.toml -v /var/run/docker.sock:/var/run/docker.sock palletone/gptn:1.0.3 --testnet

- ## 作为本地搭建私有链节点

- 克隆 go-palletone 项目， git clone -b mainnet https://github.com/palletone/go-palletone.git，进入根目录下的 examples 目录下的 first-network 目录下，查看 README.md 文件中相应的操作步骤即可

- [REDAME.md](https://github.com/palletone/go-palletone/tree/master/examples/first-network)
