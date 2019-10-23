下载中间镜像
* docker pull palletone/middle:step1

进入容器构建多节点
* docker run -it --name step2 palletone/middle:step1 bash

* cd images

* ./bytn.sh
* cd scripts
* mkdir result
* mv node* result

退出，并构建中间镜像
* docker commit step2 palletone/step2

构建最终镜像
* docker build -t palletone/private-gptn .
