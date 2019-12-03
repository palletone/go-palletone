1 install prometheus
1.1 https://prometheus.io/download/
1.2 wget https://github.com/prometheus/prometheus/releases/download/v2.14.0/prometheus-2.14.0.linux-amd64.tar.gz
1.3 tar xf prometheus-2.14.0.linux-amd64.tar.gz
1.4 mv prometheus-2.14.0.linux-amd64 /usr/local/prometheus
1.5 cp go-palletone/cmd/deployment/prometheus/prometheus.yml /usr/local/prometheus
1.6 start prometheus

2 install grafana
2.1 wget https://s3-us-west-2.amazonaws.com/grafana-releases/release/grafana_5.1.4_amd64.deb
2.2 sudo apt-get install -y adduser libfontconfig
2.3 sudo dpkg -i grafana_5.1.4_amd64.deb
2.4 service grafana start

Refer:
1 https://blog.csdn.net/hjxzb/article/details/81044583

