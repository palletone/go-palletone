#!/bin/bash

#添加docker启动配置daemon.json文件
cat > /etc/docker/daemon.json <<EOF
{
  "icc":false,
  "ip-forward":false
}
EOF

#重新加载配置文件
systemctl daemon-reload

#重启docker服务
systemctl restart docker.service

#禁止宿主机IP转发功能
cat >> /etc/sysctl.conf <<EOF
net.ipv4.ip_forward=0
EOF

#重新加载配置文件
sysctl -p /etc/sysctl.conf
