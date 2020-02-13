*** Settings ***
Library           RequestsLibrary

*** Variables ***
${host}           http://127.0.0.1:8545/
# methods
${invokeMethod}    contract_ccinvoketx
${exchangeMethod}    contract_ccinvokeToken
${createTxWithOutFee}    wallet_createTxWithOutFee
${transferPTNMethod}    wallet_transferPtn
${getBalanceMethod}    wallet_getBalance
${unlockAccountMethod}    personal_unlockAccount
${prc720ContractAddr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
# comman param
${tokenHolder}    ${null}
${Alice}          ${null}
${Bob}            ${null}
${amount}         10000
${fee}            10
${pwd}            1
${res}            ${null}
${extra}          "createtx"
${duration}       600000000
#${AliceToken}     ${null}
#${BobToken}       ${null}
${AliceTokenID}    ${null}
${BobTokenID}      ${null}
