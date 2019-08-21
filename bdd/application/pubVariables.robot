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
${createTraceability}    wallet_createTraceability
${tranferToken}    wallet_transferToken
${checkProofExistence}    wallet_getProofOfExistencesByRef
${queryToken}     wallet_getProofOfExistencesByRef
${queryMaindata}    wallet_getFileInfoByFileHash
${queryAddrHistory}    wallet_getAddrTxHistory
${queryTokenHistory}    ptn_getTokenTxHistory
${pledgeDeposit}    contract_ccinvoketx
${createToken}    contract_ccinvoketxPass
${contractQuery}    contract_ccquery
# common variables
${contractAddr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM
${contractAddr2}    PCGTta3M4t3yXu8uRgkKvaWd2d8DRijspoq
${pledgeMethod}    PledgeDeposit
${createMethod}    createToken
${symbol}         YDPOC
${str1}           The jade's information in PalletOne
${str2}           [{\"TokenID\":\"00000000000000000000000000000001\",\"MetaData\":\"metadata1\"}]
${int1}           1
${int2}           3
${strnull}        ${null}
${userAccount}    ${null}
${userAccount2}    ${null}
${userAccount3}    ${null}
${tokenHolder}    ${null}
${token}          ${null}
${maindata}       maindata
${extradata}      extradata
${reference}      A
${maindata1}      存证
${extradata1}     附加信息
${reference1}     B
${amount}         10000
${tokenid}        00000000000000000000000000000002
${fee}            1
${pwd}            1
${duration}       600000000
