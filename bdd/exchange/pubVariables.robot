*** Settings ***
Library           RequestsLibrary

*** Variables ***
${host}           http://127.0.0.1:8545/
# methods
${invokeMethod}    contract_ccinvoketx
${exchangeMethod}    contract_ccinvokeToken
${queryMethod}    contract_ccquery
${transferPTNMethod}    wallet_transferPtn
${transferTokenMethod}    wallet_transferToken
${getBalanceMethod}    wallet_getBalance
${unlockAccountMethod}    personal_unlockAccount
${personalListAccountsMethod}    personal_listAccounts
# address
${prc720ContractAddr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
${exchangeContractAddr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DS36t3ba
# comman param
${tokenHolder}    ${null}
${Alice}          ${null}
${Bob}            ${null}
${Carol}          ${null}
${amount}         10000
${fee}            1
${pwd}            1
${res}            ${null}
${duration}       600000000
${AliceToken}     BVC
${BobToken}       BOB
${CarolToken}     CAROL
${exchangesn}     ExchangeSn
${AliceTokenID}    ${null}
${BobTokenID}     ${null}
${CarolTokenID}    ${null}
