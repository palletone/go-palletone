#!/bin/bash

./test_case_setup.sh $1

sleep 1

listAccounts=`../node/gptn --exec 'personal.listAccounts' attach ../node/palletone/gptn.ipc`

key=`echo $listAccounts`
echo $key

list=`echo $key | jq ''`;

length=`echo $key | jq 'length'`

num=$[$length - 1]

for index in `seq 0 $num`
do
	if [ $index -gt 5 ]
	then
	account0=`echo $list|jq ".[0]"`
	another=`echo $list|jq ".[$index]"`
	./transfertoken.sh $account0 $another
	sleep 3 
	fi
#	#echo $list | jq ".[$index]";
done

mediatorAddr_01=`echo $list|jq ".[0]"`

#foundationAddr=`echo $list|jq ".[6]"`

../node/gptn --exec "personal.unlockAccount($mediatorAddr_01,'1',0)" attach ../node/palletone/gptn.ipc

for index in `seq 0 $num`
do
	if [ $index -gt 5 ]
	then
	toAddr=`echo $list | jq ".[$index]"`
	../node/gptn --exec "personal.unlockAccount($toAddr,'1',0)" attach ../node/palletone/gptn.ipc
	fi
	#echo $list | jq ".[$index]";
done

mdi_01=`echo ${mediatorAddr_01//\"/}`

found=`echo ${mediatorAddr_01//\"/}`

mediatorAddr_02=`echo $list | jq ".[6]"`
mdi_02=`echo ${mediatorAddr_02//\"/}`

#mdiatorAddr_03=`echo $list | jq ".[4]"`
#mdi_03=`echo ${mediatorAddr_03//\"/}`

juryAddr_01=`echo $list | jq ".[7]"`
jury_01=`echo ${juryAddr_01//\"/}`

developerAddr_01=`echo $list | jq ".[8]"`
developer_01=`echo ${developerAddr_01//\"/}`

anotherAddr=`echo $list | jq ".[9]"`
another1=`echo ${anotherAddr//\"/}`
echo $mdi_01
echo "----0000"
echo $found
pybot -d ./log -v mediatorAddr_01:$mdi_01 -v foundationAddr:$found --test Business_01 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v mediatorAddr_02:$mdi_02 -v foundationAddr:$found --test Business_02 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v juryAddr:$jury_01 --test Business_03 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v developerAddr:$developer_01 --test Business_04 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v juryAddr:$jury_01 -v foundationAddr:$found --test Business_06 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v developerAddr:$developer_01 --test Business_07 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v mediatorAddr_02:$mdi_02 --test Business_05 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v mediatorAddr_02:$mdi_02 --test Business_08 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v juryAddr:$jury_01 --test Business_09 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v juryAddr:$jury_01 -v foundationAddr:$found--test Business_10 ./deposit_test_cases/DepositContractTest.robot
#robot -d ./log -v mediatorAddr_02:$mdi_02 -v foundationAddr:$found -v anotherAddr:$another --test Business_02 ./deposit_test_cases/DepositContractTest.robot
#echo $another
#echo $mdi_02
#echo $jury_01
#echo $developer_01
#echo $another1
#pybot -d ./log -v mediatorAddr_01:$mdi_01 -v foundationAddr:$mdi_01 -v mediatorAddr_02:$mdi_02 -v juryAddr_01:$jury_01 -v developerAddr_01:$developer_01 -v anotherAddr:$another1 ./deposit_test_cases/DepositContractTest.robot


#./test_case_teardown.sh

#killall gptn
