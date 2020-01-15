1 install prometheus

1.1 https://prometheus.io/download/

1.2 wget https://github.com/prometheus/prometheus/releases/download/v2.14.0/prometheus-2.14.0.linux-amd64.tar.gz

1.3 tar xzvf prometheus-2.14.0.linux-amd64.tar.gz

1.4 mv prometheus-2.14.0.linux-amd64 /usr/local/prometheus

1.5 cp go-palletone/cmd/deployment/prometheus/prometheus.yml /usr/local/prometheus

1.6 start prometheus


2 install grafana

2.1 https://grafana.com/grafana/download

2.2 wget https://dl.grafana.com/oss/release/grafana-6.5.1.linux-amd64.tar.gz

2.3 tar xzvf grafana-6.5.1.linux-amd64.tar.gz

2.4 mv grafana-6.5.1 /usr/local/grafana

2.5 cd /usr/local/grafana/bin

2.6 start grafana-server


Refer:
1 https://blog.csdn.net/hjxzb/article/details/81044583

