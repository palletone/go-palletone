#!/bin/bash


function ModifyJson()
{
filename=../node1/ptn-genesis.json
acc=`cat $filename | jq -r '.initialMediatorCandidates[].account'`


if [ "$acc" =  "" ];then
    del=`cat $filename |  jq 'del(.initialMediatorCandidates[0])'`
    add1=`echo $del | jq ".initialMediatorCandidates[.initialMediatorCandidates| length] |= . + {\"account\": \"$1\", \"initPubKey\": \"$2\", \"node\": \"$3\"}"`

    add=`echo $add1 | 
       jq "to_entries | 
       map(if .key == \"tokenHolder\" 
          then . + {\"value\":\"$1\"} 
          else . 
          end
         ) | 
      from_entries"`

else
    #echo "is not null"
    add=`cat $filename | jq ".initialMediatorCandidates[.initialMediatorCandidates| length] |= . + {\"account\": \"$1\", \"initPubKey\": \"$2\", \"node\": \"$3\"}"`
fi

    rm $filename
    echo $add >> temp.json
    jq -r . temp.json >> $filename
    rm temp.json
#    replace $filename



}


#    {
#      "account": "P1AcP8DsqAx4Ei2gFFNLaZ16yWD6W9jo1Fh",
#      "initPubKey": "YygKiqpr81KnzKRtj2mXPHA23GXAJUn9BKoNJsJd38eFfE65SNt6x9XdMTo1xtqNUeteK1EYasfG63kB7caa8uHVSdNo4w9jF5h6hoLEyJJrvgVQNBycijj4s9FRGY9VK5nfxm7nmYZ83VgeJR4wy7dWY9bQQoqVW3emiNoTe3cqRwF",
#      "node": "pnode://1782584cdacb62be65f95ad92f9b5e997b678d38810a4b1ba8ae59fc5b9baa62bd964c3cba2e18d3686e6ae10b7ec80b4be3559f7ac443eea8f34f55ab99b657@[::]:30305"
#    }



