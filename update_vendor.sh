export GO111MODULE=on
export GOPROXY=https://goproxy.cn
go get -u github.com/palletone/adaptor@master
go get -u github.com/palletone/btc-adaptor@master
go get -u github.com/palletone/eth-adaptor@master
go get -u github.com/palletone/digital-identity@master
go mod tidy
go mod vendor