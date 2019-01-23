go get -u github.com/jteeuwen/go-bindata/...
cd internal/jsre/deps
go-bindata -nometadata -pkg deps -o bindata.go bignumber.js web3.js
gofmt -w -s bindata.go
cd ../../..