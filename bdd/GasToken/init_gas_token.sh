#!/usr/bin/env bash

## build gptn
#cd ../../
#go build ./cmd/gptn
#cp ./cmd/gptn/gptn ./bdd/GasToken/node/
#cd ./bdd/GasToken/node
#chmod +x gptn

# new genesis
./gptn newgenesis "" fasle << EOF
y
1
1
EOF

# edit genesis json
python ../pylibs/init_chain.py

# gptn init
./gptn newgenesis "" fasle << EOF
1
EOF

# start gptn
cd ../../../cmd/gptn
nohup ./gptn &