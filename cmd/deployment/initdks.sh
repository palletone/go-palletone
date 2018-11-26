
key=`./gptn mediator initdks`
echo $key

privatekeylength=45
private=${key#*private key: }
privatekey=${private:0:$privatekeylength}
echo $privatekey


publickeylength=176
public=${key#*public key: }
publickey=${public:0:$publickeylength}
echo $publickey


#key=`./gptn mediator initdks`
#privatekeylength=45
#private=${key#*private key: }
#privatekey=${private:0:$privatekeylength}
#echo $privatekey
