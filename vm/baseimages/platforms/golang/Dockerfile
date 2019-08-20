#基础镜像
FROM palletone/baseimg

#维护者信息
MAINTAINER palletOne "contract@pallet.one"

#安装相应的软件
RUN wget -o download.log https://studygolang.com/dl/golang/go1.10.linux-amd64.tar.gz \
    && tar -C /usr/local -zxvf go1.10.linux-amd64.tar.gz >> download.log \
    && rm go1.10.linux-amd64.tar.gz download.log \
    && mkdir -p /gopath/bin /gopath/src /gopath/pkg /chaincode/input

#设置环境变量
ENV GOPATH=/gopath  PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin/:$GOPATH/bin

#将宿主机文件拷贝到镜像里
COPY ./vendor /gopath/src/
