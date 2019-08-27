go get github.com/golang/mock/gomock
go install github.com/golang/mock/mockgen

mockgen -source=./ICryptoCurrency.go -aux_files="github.com/palletone/adaptor=./IUtility.go"  -destination=./CryptoCurrency_mock.go -package=adaptor -self_package="github.com/palletone/adaptor"
mockgen -source=./ISmartContract.go -aux_files="github.com/palletone/adaptor=./IUtility.go"  -destination=./SmartContract_mock.go -package=adaptor -self_package="github.com/palletone/adaptor"
mockgen -source=./IUtility.go   -destination=./Utility_mock.go -package=adaptor -self_package="github.com/palletone/adaptor"