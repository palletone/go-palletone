*** Settings ***
Default Tags      normal
Library           RequestsLibrary
Library           Collections
Library           /usr/lib/python2.7/decimal.py
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${geneAdd}        P18h3HCoFZyUsmKtMRbYqrQWdbnkiyDPNWF
${recieverAdd}    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw
${preTokenId}     QA006
${tokenDecimal}    1
${tokenAmount}    25000
${amount}         2000
${poundage}       1
${senderAmount}    2
${pdg}            1
${evidence}       evidence
${unlocktime}     ${6000000}
${contractId}     PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
${result_code}    [a-z0-9]{64}

*** Test Cases ***
transferToken_senderTokenPTN
    [Documentation]    ${preTokenId} must be a new one
    [Tags]    normal
    ${GeneAdd}    getGeneAdd    ${host}
    normalCcinvokePass    ${result_code}    ${pretokenId}    ${tokenDecimal}    ${tokenAmount}    ${amount}    ${poundage}
    sleep    5
    ${result1}    getBalance    ${GeneAdd}
    sleep    5
    ${key}    getTokenId    ${preTokenId}    ${result1}
    ${PTN1}    Get From Dictionary    ${result1}    PTN
    ${item1}    Get From Dictionary    ${result1}    ${key}
    ${tokenResult}    transferToken    ${key}    ${GeneAdd}    ${recieverAdd}    ${senderAmount}    ${pdg}
    ...    ${evidence}    ${unlocktime}
    ${item'}    Evaluate    ${item1}-${senderAmount}
    ${PTN'}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${pdg}')    decimal
    sleep    5
    ${result2}    getBalance    ${GeneAdd}
    ${PTN2}    Get From Dictionary    ${result2}    PTN
    ${item2}    Get From Dictionary    ${result2}    ${key}
    Should Be Equal As Numbers    ${item2}    ${item'}
    Should Be Equal As Strings    ${PTN2}    ${PTN'}
