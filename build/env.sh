#!/bin/sh

set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

go get -u github.com/palletone/btc-adaptor
go get -u github.com/palletone/eth-adaptor


# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
ethdir="$workspace/src/github.com/palletone"
if [ ! -L "$ethdir/go-palletone" ]; then
    mkdir -p "$ethdir"
    cd "$ethdir"
    ln -s ../../../../../. go-palletone
    ln -s ../../../../../../btc-adaptor/. btc-adaptor
    ln -s ../../../../../../eth-adaptor/. eth-adaptor

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
