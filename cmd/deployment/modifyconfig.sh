
newipcpath="IPCPath=\"gptn1.ipc\""
sed -i '/^IPCPath/c'$newipcpath'' ptn-config.toml

newHTTPPort="HTTPPort=8555"
sed -i '/^HTTPPort/c'$newHTTPPort'' ptn-config.toml

newWSPort="WSPort=8556"
sed -i '/^WSPort/c'$newWSPort'' ptn-config.toml

newPort="Port=8081"
sed -i '/^Port/c'$newPort'' ptn-config.toml


newListenAddr="ListenAddr=\":30305\""
sed -i '/^ListenAddr/c'$newListenAddr'' ptn-config.toml

newBtcHost="BtcHost=\"localhost:18333\""
sed -i '/^BtcHost/c'$newBtcHost'' ptn-config.toml


newContractAddress="ContractAddress=\"127.0.0.1:12346\""
sed -i '/^ContractAddress/c'$newContractAddress'' ptn-config.toml

createaccount=`./createaccount.sh`
#echo "...\r $createaccount"
account=`echo $createaccount | sed -n '$p'| awk '{print $NF}'`
echo "account: "$account




key=`./gptn mediator initdks`

privatekeylength=45
private=${key#*private key: }
privatekey=${private:0:$privatekeylength}
echo $privatekey


publickeylength=176
public=${key#*public key: }
publickey=${public:0:$publickeylength}
echo $publickey


newAddress="Address=\"$account\""
sed -i '/^Address/c'$newAddress'' ptn-config.toml


newPassword="Password=\"1\""
sed -i '/^Password/c'$newPassword'' ptn-config.toml

newInitPartSec="InitPartSec=\"$privatekey\""
sed -i '/^InitPartSec/c'$newInitPartSec'' ptn-config.toml


newInitPartPub="InitPartPub=\"$publickey\""
sed -i '/^InitPartPub/c'$newInitPartPub'' ptn-config.toml







