*** Settings ***
Library           RequestsLibrary

*** Variables ***
${host}           http://localhost:8600/
# methods
${transerferPTNMethod}    wallet_transferPtn
${getBalanceMethod}    wallet_getBalance
${personalListAccountsMethod}    personal_listAccounts
${unlockAccountMethod}    personal_unlockAccount
${createProofExistence}    wallet_createProofOfExistenceTx
${checkProofExistence}    wallet_getProofOfExistencesByRef
# common variables
${userAccount}    ${null}
${tokenHolder}    ${null}
${maindata}       maindata
${extradata}      extradata
${reference}      A
${amount}         10000
${fee}            1
${pwd}            1
${duration}       600000000
