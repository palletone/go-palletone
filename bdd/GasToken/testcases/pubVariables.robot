*** Settings ***
Library           RequestsLibrary

*** Variables ***
${host}           http://localhost:8555/
# methods
${invokeMethod}    contract_ccinvoketx
${queryMethod}    contract_ccquery
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
