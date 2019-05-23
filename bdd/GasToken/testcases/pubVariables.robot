*** Settings ***
Library           RequestsLibrary

*** Variables ***
${host}           http://localhost:8545/
# methods
${invokeMethod}    contract_ccinvoketx
${queryMethod}    contract_ccquery
${transferPTNMethod}    wallet_transferPtn
${getBalanceMethod}    wallet_getBalance
${unlockAccountMethod}    personal_unlockAccount
${personalListAccountsMethod}    personal_listAccounts
# address
# comman param
${tokenHolder}    ${null}
${AliceAddr}      ${null}
${BobAddr}        ${null}
${CarolAddr}      ${null}
${amount}         10000
${fee}            1
${pwd}            1
${duration}       600000000
${gasToken}       WWW
