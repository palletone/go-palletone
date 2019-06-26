*** Settings ***
Library           RequestsLibrary

*** Variables ***
${host}           http://123.126.106.82:58645/
${juryHosts}      Create List
#${host}          http://localhost:8545/
${gastokenHost}    http://localhost:8545
${nodenum}        3
# methods
${invokeMethod}    contract_ccinvoketx
${queryMethod}    contract_ccquery
${installMethod}    contract_ccinstalltx
${deployMethod}    contract_ccdeploytx
${transferPTNMethod}    wallet_transferPtn
${transferTokenMethod}    wallet_transferToken
${getBalanceMethod}    wallet_getBalance
${unlockAccountMethod}    personal_unlockAccount
${personalListAccountsMethod}    personal_listAccounts
# address
${prc720ContractAddr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
# comman param
${tokenHolder}    ${null}
${Alice}          ${null}
${Bob}            ${null}
${Carol}          ${null}
${amount}         10000
${fee}            1
${pwd}            1
${duration}       600000000
${gasToken}       WWW
${AliceToken}     ALICE
${BobToken}       BOB
${CarolToken}     CAROL
${AliceTokenID}    ${null}
${BobTokenID}     ${null}
${CarolTokenID}    ${null}
