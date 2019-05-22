#!/bin/sh

set -e
if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi
ada=github.com/palletone/adaptor
btc=github.com/palletone/btc-adaptor
eth=github.com/palletone/eth-adaptor
di=github.com/palletone/digital-identity

adafullpath=$GOPATH/src/$ada
btcfullpath=$GOPATH/src/$btc
ethfullpath=$GOPATH/src/$eth
difullpath=$GOPATH/src/$di

echo $adafullpath
echo $btcfullpath
echo $ethfullpath
echo $difullpath


#go get -u github.com/palletone/eth-adaptor

if [ ! -d "$adafullpath" ]; then
	echo "adaptor not exist"
	go get -u $ada
else
	echo "adaptor exist"
#	git -C $adafullpath  pull
fi


if [ ! -d "$btcfullpath" ]; then
	echo "btc not exist"
	go get -u $btc
else
	echo "btc exist"
#	git -C $btcfullpath pull
fi


if [ ! -d "$ethfullpath" ]; then
	echo "eth not exist"
	go get -u $eth
else
	echo "eth exist"
#	git -C $ethfullpath  pull
fi


if [ ! -d "$difullpath" ]; then
	echo "di not exist"
	go get -u $di
else
	echo "di exist"
#	git -C $difullpath  pull
fi



# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
ethdir="$workspace/src/github.com/palletone"
if [ ! -L "$ethdir/go-palletone" ]; then
    mkdir -p "$ethdir"
    cd "$ethdir"
    ln -s ../../../../../. go-palletone
    ln -s ../../../../../../adaptor/. adaptor
    ln -s ../../../../../../btc-adaptor/. btc-adaptor
    ln -s ../../../../../../eth-adaptor/. eth-adaptor
    ln -s ../../../../../../digital-identity/. digital-identity

    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH

# Run the command inside the workspace.
cd "$ethdir/go-palletone"
PWD="$ethdir/go-palletone"

# Launch the arguments with the configured environment.
exec "$@"
