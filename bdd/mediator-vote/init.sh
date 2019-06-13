#!/bin/bash
#pkill gptn
#tskill gptn
#cd ../../cmd/gptn && go build
cd ../../
#rm -rf ./bdd/mediator-vote/node*
#cp ./cmd/gptn/gptn ./bdd/mediator-vote/node
cd ./bdd/mediator-vote/node
chmod +x gptn

# new genesis
./gptn newgenesis "" fasle << EOF
y
1
1
EOF

# edit genesis json
jsonFile="ptn-genesis.json"
if [ -e "$jsonFile" ]; then
    #file already exist, modify
    sed -i "s/\"activeMediatorCount\": \"5\"/\"activeMediatorCount\": \"3\"/g" $jsonFile
    sed -i "s/\"initialActiveMediators\": \"5\"/\"initialActiveMediators\": \"3\"/g" $jsonFile
    sed -i "s/\"minMediatorCount\": \"5\"/\"minMediatorCount\": \"3\"/g" $jsonFile
    sed -i "s/\"maintenanceInterval\": \"600\"/\"maintenanceInterval\": \"150\"/g" $jsonFile
else
    #file not found, new file
    echo "no $jsonFile"
    exit -1
fi

# edit toml file
tomlFile="ptn-config.toml"
if [ -e "$tomlFile" ]; then
    #file already exist, modify
    sed -i "s/HTTPPort = 8545/HTTPPort = 8595/g" $tomlFile
    sed -i "s/WSPort = 8546/WSPort = 8596/g" $tomlFile
    sed -i "s/Port = 8080/Port = 8091/g" $tomlFile
    sed -i "s/ListenAddr = \":30303\"/ListenAddr = \":30393\"/g" $tomlFile
    sed -i "s/CorsListenAddr = \":50505\"/CorsListenAddr = \":50595\"/g" $tomlFile
    sed -i "s/BtcHost = \"localhost:18332\"/BtcHost = \"localhost:18392\"/g" $tomlFile
    sed -i "s/ContractAddress = \"127.0.0.1:12345\"/ContractAddress = \"127.0.0.1:12395\"/g" $tomlFile
    sed -i "s/CaUrl = \"http://localhost:8545\"/CaUrl = \"http://localhost:8595\"/g" $tomlFile
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