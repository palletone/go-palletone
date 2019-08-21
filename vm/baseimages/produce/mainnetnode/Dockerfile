#基础镜像
FROM ubuntu:18.04

#维护者信息
MAINTAINER palletone "contract@pallet.one"

#将宿主机压缩文件添加到镜像里
ADD ./tardata/* /palletone/

#设置当前工作目录
WORKDIR /palletone/mainnet

#容器启动执行命令
CMD ["./gptn"]
