*** Settings ***
Library           RequestsLibrary
Library           Collections
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${method}         ptn_signRawTransaction
${geneAdd}        P1CwGYGSjWSaJrysHAjAWtDyFSsbcYwoULv
${recieverAdd}    P11LKXXsDuKUdo3cJEy7uMnqnvw4kwHPs8
@{transList}      ${geneAdd}    ${recieverAdd}    10    2
#${crtResult}     f8bff8bdf8bb80b8b8f8b6e6e58080a005ba8746c65994e7c851ce81aa437f5856572c6f1d67533e5985ceac9a7dcad18064f88cf842843b9aca009976a91488be35cc7c2cace89a62193d96cb45e93d808ec688ace290400082bb0800000000000000000000009000000000000000000000000000000000f84688015f800022f0c5009976a914f1afa0fcca3d943a0092c0f1896ad09c3933074b88ace290400082bb080000000000000000000000900000000000000000000000000000000080
${result_code}    \f[a-z0-9]*
${result_hex}     \f[a-z0-9]*
${result_txid}    \0[a-z0-9]{160,170}

*** Test Cases ***
signTransNormal
    [Tags]    normal
    [Template]    normalSignTrans
    ${result_hex}    ${result_txid}

*** Keywords ***
