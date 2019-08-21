#基础镜像
FROM ubuntu:18.04

#维护者信息
MAINTAINER palletone "contract@pallet.one"

#复制宿主机文件到镜像
COPY /node1/gptn /palletone/mediator/

#为后续指令设置工作目录
WORKDIR /palletone/mediator

#启动容器时执行的命令
CMD ["./gptn"]
