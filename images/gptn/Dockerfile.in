# Copyright <dev@pallet.one> All Rights Reserved
#
#
#基础镜像
FROM _BASE_NS_/pallet-baseos:_BASE_TAG_
#创建目录
RUN mkdir -p /var/palletone/production /var/palletone/conf /var/palletone/log
#拷贝可执行文件
COPY ./gptn /usr/local/bin
#拷贝主执行文件
COPY ./entrypoint.sh /var/palletone/conf
#赋予权限
RUN chmod a+x /usr/local/bin/gptn /var/palletone/conf/entrypoint.sh
#开放容器对外端口
EXPOSE 8545 8546 8080 30303 18332 12345
#启动命令 
ENTRYPOINT ["/var/palletone/conf/entrypoint.sh"]
