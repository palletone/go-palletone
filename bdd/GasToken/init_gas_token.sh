#!/bin/bash
#pkill gptn
#tskill gptn
#cd ../../cmd/gptn && go build
cd ../../
#rm -rf ./bdd/GasToken/node/*
#cp ./cmd/gptn/gptn ./bdd/GasToken/node
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
    sed -i 's/"mediator_interval": 3,/"mediator_interval": 2,/g' $jsonFile
    sed -i 's/"initialTimestamp": [0-9]*,/"initialTimestamp": 1566269000,/g' $jsonFile
    sed -i 's/"maintenance_skip_slots": 1,/"maintenance_skip_slots": 0,/g' $jsonFile
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
    sed -i "s/HTTPPort = 8545/HTTPPort = 8555/g" $tomlFile
    sed -i "s/WSPort = 8546/WSPort = 8556/g" $tomlFile
    sed -i "s/Port = 8080/Port = 8090/g" $tomlFile
    sed -i "s/ListenAddr = \":30303\"/ListenAddr = \":30313\"/g" $tomlFile
    sed -i "s/CorsListenAddr = \":50505\"/CorsListenAddr = \":50515\"/g" $tomlFile
    sed -i "s/BtcHost = \"localhost:18332\"/BtcHost = \"localhost:18342\"/g" $tomlFile
    sed -i "s/ContractAddress = \"127.0.0.1:12345\"/ContractAddress = \"127.0.0.1:12355\"/g" $tomlFile
    sed -i "s/CaUrl = \"http://localhost:8545\"/CaUrl = \"http://localhost:8555\"/g" $tomlFile
    sed -i "s/OutputPaths = \[\"stdout\", \".\/log\/all.log\"\]/OutputPaths = \[\".\/log\/all.log\"\]/g" $tomlFile
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