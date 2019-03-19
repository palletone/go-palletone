#!/bin/bash

./test_case_setup.sh $1

listAccounts=`./node1/gptn --exec 'personal.listAccounts' attach node1/palletone/gptn.ipc`

key=`echo $listAccounts`
#echo $key

list=`echo $key | jq ''`;

length=`echo $key | jq 'length'`

num=$[$length - 1]

for index in `seq 0 $num`
do
	if [ $index -gt 1 ]
	then
	account0=`echo $list|jq ".[1]"`
	another=`echo $list|jq ".[$index]"`
	./transfertoken.sh $account0 $another
	sleep 3 
	fi
	#echo $list | jq ".[$index]";
done

mediatorAddr_01=`echo $list|jq ".[1]"`

foundationAddr=`echo $list|jq ".[2]"`

./node1/gptn --exec "personal.unlockAccount($mediatorAddr_01,'1',0)" attach node1/palletone/gptn.ipc

for index in `seq 0 $num`
do
	if [ $index -gt 1 ]
	then
	toAddr=`echo $list | jq ".[$index]"`
	./node1/gptn --exec "personal.unlockAccount($toAddr,'1',0)" attach node1/palletone/gptn.ipc
	fi
	#echo $list | jq ".[$index]";
done

mdi_01=`echo ${mediatorAddr_01//\"/}`

found=`echo ${foundationAddr//\"/}`

mediatorAddr_02=`echo $list | jq ".[3]"`
mdi_02=`echo ${mediatorAddr_02//\"/}`

#mdiatorAddr_03=`echo $list | jq ".[4]"`
#mdi_03=`echo ${mediatorAddr_03//\"/}`

juryAddr_01=`echo $list | jq ".[5]"`
jury_01=`echo ${juryAddr_01//\"/}`

developerAddr_01=`echo $list | jq ".[6]"`
developer_01=`echo ${developerAddr_01//\"/}`

anotherAddr=`echo $list | jq ".[4]"`
another=`echo ${anotherAddr//\"/}`

#robot -d ./log -v mediatorAddr_01:$mdi_01 -v foundationAddr:$found --test Business_01 ./deposit_test_cases/DepositContractTest.robot
#robot -d ./log -v mediatorAddr_02:$mdi_02 -v foundationAddr:$found -v anotherAddr:$another --test Business_02 ./deposit_test_cases/DepositContractTest.robot
#echo $another
robot -d ./log -v mediatorAddr_01:$mdi_01 -v foundationAddr:$found -v mediatorAddr_02:$mdi_02 -v juryAddr_01:$jury_01 -v developerAddr_01:$developer_01 -v anotherAddr:$another ./deposit_test_cases/DepositContractTest.robot


#./test_case_teardown.sh

killall gptn
