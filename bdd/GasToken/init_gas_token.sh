#!/usr/bin/env bash
cd ../../
## build gptn
go build ./cmd/gptn
cp ./cmd/gptn/gptn ./bdd/GasToken/node/
cd ./bdd/GasToken/node
chmod +x gptn

# new genesis
./gptn newgenesis "" fasle << EOF
y
1
1
EOF

# edit genesis json
gasToken="WWW"
jsonFile="ptn-genesis.json"
if [ -e "$jsonFile" ]; then
    #file already exist, modify
    sed -i "s/\"gasToken\": \"PTN\"/\"gasToken\": \"$gasToken\"/g" $jsonFile
else
    #file not found, new file
    echo "no $jsonFile"
    exit -1
fi

# edit toml file
tomlFile="ptn-config.toml"
if [ -e "$tomlFile" ]; then
    #file already exist, modify
    sed -i "s/GasToken = \"PTN\"/GasToken = \"$gasToken\"/g" $tomlFile
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