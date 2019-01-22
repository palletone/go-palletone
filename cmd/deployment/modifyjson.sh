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

index=$[ $4 - 1 ]

add=`cat $filename | jq ".initialMediatorCandidates[$index] |= . + {\"account\": \"$1\", \"initPubKey\": \"$2\", \"node\": \"$3\"}"`

if [ $index -eq 0 ] ; then

    createaccount=`./createaccount.sh`
    tempinfo=`echo $createaccount | sed -n '$p'| awk '{print $NF}'`
    accountlength=35
    accounttemp=${tempinfo:0:$accountlength}
    account=`echo ${accounttemp//^M/}`

    add=`echo $add |
       jq "to_entries |
       map(if .key == \"tokenHolder\"
          then . + {\"value\":\"$account\"}
          else .
          end
         ) |
      from_entries"`

fi

    rm $filename
    echo $add >> temp.json
    jq -r . temp.json >> $filename
    rm temp.json
}


