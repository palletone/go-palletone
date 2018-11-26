#json=`./gptn dumpjson`
#nodeinfo=`./gptn nodeInfo`
#info=`echo $nodeinfo | sed -n '$p'| awk '{print $NF}'`
#echo "nodeinfo: "$info
nodeinfo="pnode://1782584cdacb62be65f95ad92f9b5e997b678d38810a4b1ba8ae59fc5b9baa62bd964c3cba2e18d3686e6ae10b7ec80b4be3559f7ac443eea8f34f55ab99b657@[::]:30305"

account="P17AK35UVKd91y6YL4kxNpwvd86THhuL45B"
publickey="ggVU59QexbJeNQC9Qk15jpNBAkVfEfCXH2XepNMgtNXqSHM7B2cgFSfU9ngLQgkKJRFs1uxtuzZdT4QYPuFqnimKb3tLRC6hW42TkJTfcnWLnLi7PBxRaSjnTeyYRYG6xcsQYvjxS8M1dyxyojV4Cf1kEYNWDymPGHhCVP5NAZGNhbr"

#initialMediatorCandidates

#jq '{"time"}?'
filename=ptn-genesis.json


#echo ${complexJson} |jq -r '.nameInfo[].lastName'
acc=`cat $filename | jq -r '.initialMediatorCandidates[].account'`

echo "...account:"$acc

if [ "$acc" =  "" ];then
    echo "is null"
    del=`cat $filename |  jq 'del(.initialMediatorCandidates[0])'`
    #jq '.data.messages += [{"date": "2010-01-07T19:55:99.999Z", "xml": "xml_samplesheet_2017_01_07_run_09.xml", "status": "OKKK", "message": "metadata loaded into iRODS successfullyyyyy"}]' 
    add=`echo $del | jq ".initialMediatorCandidates[.initialMediatorCandidates| length] |= . + {\"account\": \"$account\", \"initPubKey\": \"$publickey\", \"node\": \"$nodeinfo\"}"`
    rm temp.json genesis.json
    echo $add >> temp.json
    jq -r . temp.json >> genesis.json
    rm temp.json


else
    echo "is not null"
fi


#    {
#      "account": "P1AcP8DsqAx4Ei2gFFNLaZ16yWD6W9jo1Fh",
#      "initPubKey": "YygKiqpr81KnzKRtj2mXPHA23GXAJUn9BKoNJsJd38eFfE65SNt6x9XdMTo1xtqNUeteK1EYasfG63kB7caa8uHVSdNo4w9jF5h6hoLEyJJrvgVQNBycijj4s9FRGY9VK5nfxm7nmYZ83VgeJR4wy7dWY9bQQoqVW3emiNoTe3cqRwF",
#      "node": "pnode://1782584cdacb62be65f95ad92f9b5e997b678d38810a4b1ba8ae59fc5b9baa62bd964c3cba2e18d3686e6ae10b7ec80b4be3559f7ac443eea8f34f55ab99b657@[::]:30305"
#    }



