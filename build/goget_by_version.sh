function goversion()
{
    version=$(go version)
    regex="go([0-9]+.[0-9]+).[0-9]+"
    if [[ $version =~ $regex ]];
    then
         echo ${BASH_REMATCH[1]};
    else
         echo "Not proper format";
    fi
}
if [ $(goversion) == "1.12" ] ;
then
        go get -u github.com/palletone/digital-identity
        go get github.com/golang/mock/gomock
        go install github.com/golang/mock/mockgen
else
        go get -u github.com/palletone/digital-identity/...
        mkdir -p $GOPATH/src/github.com/golang
        cd $GOPATH/src/github.com/golang
        git clone https://github.com/golang/mock.git
        cd mock
        git checkout 442550a
        go install github.com/golang/mock/mockgen
        cd $GOPATH/src/github.com/palletone/go-palletone
fi