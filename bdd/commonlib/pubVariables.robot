*** Settings ***
Library           RequestsLibrary

*** Variables ***
#${ip}            123.126.106.82
#${host}          http://${ip}:58645/
${ip}             127.0.0.1
${host}           http://${ip}:8645/
${juryHosts}      Create List
${gastokenHost}    http://localhost:8545
${nodenum}        3
# methods
${ccinvokeMethod}    contract_ccinvoketx
${ccqueryMethod}    contract_ccquery
${ccinstallMethod}    contract_ccinstalltx
${ccdeployMethod}    contract_ccdeploytx
${ccstopMethod}    contract_ccstoptx
${transferPTNMethod}    wallet_transferPtn
${transferTokenMethod}    wallet_transferToken
${getBalanceMethod}    wallet_getBalance
${unlockAccountMethod}    personal_unlockAccount
${personalListAccountsMethod}    personal_listAccounts
# address
${prc720ContractAddr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
# comman param
${tokenHolder}    P128sz4bD9akVYtvnx3er3hWdUAz8nALox9
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
