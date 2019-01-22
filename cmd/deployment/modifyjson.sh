#!/bin/bash


function ModifyJson()
{
filename=../node1/ptn-genesis.json

:<<!
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
    add=`cat $filename | jq ".initialMediatorCandidates[.initialMediatorCandidates| length] |= . + {\"account\": \"$1\", \"initPubKey\": \"$2\", \"node\": \"$3\"}"`
fi
!

add=`cat $filename | jq ".initialMediatorCandidates[$4-1] |= . + {\"account\": \"$1\", \"initPubKey\": \"$2\", \"node\": \"$3\"}"`

    rm $filename
    echo $add >> temp.json
    jq -r . temp.json >> $filename
    rm temp.json
}




