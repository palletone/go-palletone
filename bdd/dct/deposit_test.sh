#!/bin/bash

./test_case_setup.sh $1

sleep 1

listAccounts=`../node/gptn --exec 'personal.listAccounts' attach ../node/palletone/gptn.ipc`
key=`echo $listAccounts `
echo $key

list=`echo $key | jq ''`;

length=`echo $key | jq 'length'`

num=$[$length - $1]
echo "num=" $num
addrs=$[$length - $num]

echo "addrs=" $addrs

foundationAddr=`echo $list|jq ".[0]"`
found=`echo ${foundationAddr//\"/}`

all=$[$addrs - 1]
for index in `seq 0 $all`
do
	another=`echo $list|jq ".[$num + $index]"`
	./transfertoken.sh $foundationAddr $another
	sleep 3
#	#echo $list | jq ".[$index]";
done

../node/gptn --exec "personal.unlockAccount($foundationAddr,'1',0)" attach ../node/palletone/gptn.ipc

for index in `seq 0 $all`
do
	toAddr=`echo $list | jq ".[$num + $index]"`
	../node/gptn --exec "personal.unlockAccount($toAddr,'1',0)" attach ../node/palletone/gptn.ipc
done

mediatorAddr_02=`echo $list | jq ".[$num]"`
mdi_02=`echo ${mediatorAddr_02//\"/}`

num2=$[$num + 1]
juryAddr_01=`echo $list | jq ".[$num2]"`
jury_01=`echo ${juryAddr_01//\"/}`

#获取陪审团节点的公钥
jury_01_pubkey=`../node/gptn --exec "personal.getPublicKey($juryAddr_01)" attach ../node/palletone/gptn.ipc`
jury_01_pub=`echo ${jury_01_pubkey//\"/}`

num3=$[$num + 2]
developerAddr_01=`echo $list | jq ".[$num3]"`
developer_01=`echo ${developerAddr_01//\"/}`

num4=$[$num + 3]
juryA_02=`echo $list | jq ".[$num4]"`
jury_02=`echo ${juryA_02//\"/}`
#获取陪审团节点的公钥
jury_02_pubkey=`../node/gptn --exec "personal.getPublicKey($juryA_02)" attach ../node/palletone/gptn.ipc`
jury_02_pub=`echo ${jury_02_pubkey//\"/}`

num5=$[$num + 4]
deveA_02=`echo $list | jq ".[$num5]"`
developer_02=`echo ${deveA_02//\"/}`

num6=$[$num + 5]
m01=`echo $list | jq ".[$num6]"`
mdi_01=`echo ${m01//\"/}`
m1=`../node/gptn --exec "personal.getPublicKey($m01)" attach ../node/palletone/gptn.ipc`
m11=`echo ${m1//\"/}`

m2=`../node/gptn --exec "personal.getPublicKey($mediatorAddr_02)" attach ../node/palletone/gptn.ipc`
m22=`echo ${m2//\"/}`

num7=$[$num + 6]
votedAddr=`echo $list | jq ".[$num7]"`
votedAddress=`echo ${votedAddr//\"/}`

num8=$[$num + 7]
vote01=`echo $list | jq ".[$num8]"`
votedAddress01=`echo ${vote01//\"/}`

num9=$[$num + 8]
vote02=`echo $list | jq ".[$num9]"`
votedAddress02=`echo ${vote02//\"/}`

num10=$[$num + 9]
vote03=`echo $list | jq ".[$num10]"`
votedAddress03=`echo ${vote03//\"/}`

num11=$[$num + 10]
vote04=`echo $list | jq ".[$num11]"`
votedAddress04=`echo ${vote04//\"/}`

num12=$[$num + 11]
vote05=`echo $list | jq ".[$num12]"`
votedAddress05=`echo ${vote05//\"/}`

num13=$[$num + 12]
vote06=`echo $list | jq ".[$num13]"`
votedAddress06=`echo ${vote06//\"/}`

echo "found" $found
echo "mdi_01" $mdi_01
echo "mdi_01_pubkey =>" $m11
echo "mdi_02" $mdi_02
echo "mdi_02_pubkey =>" $m22
echo "juryA_01" $jury_01
echo "jury_01_pubkey =>"  $jury_01_pub
echo "juryA_02" $jury_02
echo "jury_02_pubkey =>"  $jury_02_pub
echo "devA_01" $developer_01
echo "dev_02" $developer_02
echo "vote" $votedAddress
echo "vote01" $votedAddress01
echo "vote02" $votedAddress02
echo "vote03" $votedAddress03
echo "vote04" $votedAddress04
echo "vote05" $votedAddress05
echo "vote06" $votedAddress06
#pybot -d ./log -v mediatorAddr_01:$mdi_01 -v foundationAddr:$found --test Business_01 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v mediatorAddr_02:$mdi_02 -v foundationAddr:$found --test Business_02 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v juryAddr_01:$jury_01 -v foundationAddr:$found --test Business_03 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v developerAddr_01:$developer_01 -v foundationAddr:$found --test Business_04 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v juryAddr_02:$jury_02 -v foundationAddr:$found --test Business_05 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v developerAddr_02:$developer_02 -v foundationAddr:$found --test Business_06 ./deposit_test_cases/DepositContractTest.robot
pybot -d ../logs/deposit -v votedAddress02:$votedAddress02 -v votedAddress03:$votedAddress03 -v votedAddress04:$votedAddress04 -v votedAddress05:$votedAddress05 -v votedAddress06:$votedAddress06 -v votedAddress01:$votedAddress01 -v votedAddress:$votedAddress -v m1_pubkey:$m11 -v m2_pubkey:$m22 -v juryAddr_01_pubkey:$jury_01_pub -v juryAddr_02_pubkey:$jury_02_pub -v mediatorAddr_01:$mdi_01 -v foundationAddr:$found -v mediatorAddr_02:$mdi_02 -v juryAddr_01:$jury_01 -v developerAddr_01:$developer_01 -v juryAddr_02:$jury_02 -v developerAddr_02:$developer_02 ./deposit_test_cases/DepositContractTest.robot
###pybot -d ../logs/deposit -v votedAddress02:$votedAddress02 -v votedAddress03:$votedAddress03 -v votedAddress04:$votedAddress04 -v votedAddress05:$votedAddress05 -v votedAddress06:$votedAddress06 -v votedAddress01:$votedAddress01 -v votedAddress:$votedAddress -v m1_pubkey:$m11 -v m2_pubkey:$m22 -v juryAddr_01_pubkey:$jury_01_pub -v juryAddr_02_pubkey:$jury_02_pub -v mediatorAddr_01:$mdi_01 -v foundationAddr:$found -v mediatorAddr_02:$mdi_02 -v juryAddr_01:$jury_01 -v developerAddr_01:$developer_01 -v juryAddr_02:$jury_02 -v developerAddr_02:$developer_02 --test PledgeTest02 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ../logs/deposit -v votedAddress:$votedAddress -v m1_pubkey:$m11 -v m2_pubkey:$m22 -v juryAddr_01_pubkey:$jury_01_pub -v juryAddr_02_pubkey:$jury_02_pub -v mediatorAddr_01:$mdi_01 -v foundationAddr:$found -v mediatorAddr_02:$mdi_02 -v juryAddr_01:$jury_01 -v developerAddr_01:$developer_01 -v juryAddr_02:$jury_02 -v developerAddr_02:$developer_02 --test PledgeTest04 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v mediatorAddr_01:$mdi_01 -v foundationAddr:$found -v mediatorAddr_02:$mdi_02 -v juryAddr_01:$jury_01 -v developerAddr_01:$developer_01 -v juryAddr_02:$jury_02 -v developerAddr_02:$developer_02 --test Business_01 --test Business_03 --test Business_05 ./deposit_test_cases/DepositContractTest.robot
#pybot -d ./log -v mediatorAddr_01:$mdi_01 -v foundationAddr:$found -v otherAddr:$otherAddr --test Business_07 ./deposit_test_cases/DepositContractTest.robot

#pybot -d ./log -v foundationAddr:$found  -v votedAddress:$votedAddress --test PledgeTest ./deposit_test_cases/DepositContractTest.robot

#./test_case_teardown.sh

#killall gptn
