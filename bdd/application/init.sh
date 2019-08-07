#!/bin/bash
#pkill gptn
#tskill gptn
#cd ../../cmd/gptn && go build
cd ../../
#rm -rf ./bdd/application/node*
#cp ./cmd/gptn/gptn ./bdd/application/node
cd ./bdd/application/node
chmod +x gptn

# new genesis
./gptn newgenesis "" fasle << EOF
y
1
1
EOF

# edit toml file
tomlFile="ptn-config.toml"
if [ -e "$tomlFile" ]; then
    #file already exist, modify
    sed -i "s/HTTPPort = 8545/HTTPPort = 8600/g" $tomlFile
    sed -i "s/AddrTxsIndex = false/AddrTxsIndex = true/g" $tomlFile
else
    #file not found, new file
    echo "no $tomlFile"
    exit -1
fi

# gptn init
./gptn init << EOF
1
EOF

# start gptn
nohup ./gptn &
