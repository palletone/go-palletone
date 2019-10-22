# Supported tags and respective Dockerfile links

- [latest](https://github.com/palletone/go-palletone/blob/master/images/node/Dockerfile)
- [1.0.2](https://github.com/palletone/go-palletone/blob/master/images/node/Dockerfile)
- [1.0.1](https://github.com/palletone/go-palletone/blob/master/images/node/Dockerfile)

# Quick reference

# What is node

node is used as the base image for the go-palletone node,such as the main network,test network,and the local private network...    

# How to use this image

- ### 作为主网节点

- ### 普通全节点

- 启动容器：docker run -d --name mainnetnode palletone/node:1.0.2

- 进入容器：docker exec -it mainnetnode sh

- 再进入gptn控制台：./gptn attach palletone/gptn.ipc

- ### 超级节点或者陪审员节点且不需要挂载文件

- 启动容器：docker run -d -p 8545:8545 -p 30303:30303 --network gptn-net --name mainnetnode palletone/node:1.0.2

- 进入容器：docker exec -it mainnetnode sh

- 再进入gptn控制台：./gptn attach palletone/gptn.ipc

- ### 超级节点或者陪审员节点且需要挂载文件

- 启动容器：docker run -d -p 8545:8545 -p 30303:30303 --network gptn-net --name mainnetnode -v host_absolute_path/palletone:/palletone/mainnet/palletone -v host_absolute_path/ptn-config.toml:/palletone/mainnet/ptn-config.toml palletone/node:1.0.2

- 进入容器：docker exec -it mainnetnode sh

- 再进入gptn控制台：./gptn attach palletone/gptn.ipc    

  ------

- ## 作为测试网节点

- ### 普通全节点

- 启动容器：docker run -d --name testnetnode palletone/node:1.0.2 --testnet

- 进入容器：docker exec -it testnetnode sh

- 再进入gptn控制台：./gptn attach palletone/testnet/gptn.ipc

- ### 超级节点或者陪审员节点且不需要挂载文件

- 启动容器：docker run -d -p 8545:8545 -p 30303:30303 --network gptn-net --name testnetnode palletone/node:1.0.2 --testnet

- 进入容器：docker exec -it testnetnode sh

- 再进入gptn控制台：./gptn attach palletone/testnet/gptn.ipc

- ### 超级节点或者陪审员节点且需要挂载文件

- 启动容器：docker run -d -p 8545:8545 -p 30303:30303 --network gptn-net --name testnetnode  -v host_absolute_path/palletone:/palletone/mainnet/palletone -v host_absolute_path/ptn-config.toml:/palletone/mainnet/ptn-config.toml palletone/node:1.0.2 --testnet

- 进入容器：docker exec -it testnetnode sh

- 再进入gptn控制台：./gptn attach palletone/testnet/gptn.ipc

- ## 作为本地搭建私有链节点

- 克隆 [https://github.com/palletone/go-palletone.git](https://github.com/palletone/go-palletone.git) 到本地并切换到 mainnet 分支，进入根目录下的 images 目录下，查看 README.md 文件中相应的操作步骤即可

- [REDAME.md](https://github.com/palletone/go-palletone/tree/master/images)