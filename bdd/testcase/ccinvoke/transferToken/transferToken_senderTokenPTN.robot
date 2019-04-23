*** Settings ***
Library           RequestsLibrary
Library           Collections
Library           /opt/python/2.7.15/lib/python2.7/decimal.py
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${geneAdd}        P18h3HCoFZyUsmKtMRbYqrQWdbnkiyDPNWF
${recieverAdd}    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw
${preTokenId}     QA007
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
    [Tags]    normal
    ${GeneAdd}    getGeneAdd    ${host}
    normalCcinvokePass    ${result_code}    ${pretokenId}    ${tokenDecimal}    ${tokenAmount}    ${amount}    ${poundage}
    sleep    2
    ${result1}    getBalance    ${GeneAdd}
    ${key}    getTokenId    ${preTokenId}    ${result1}
    ${PTN1}    Get From Dictionary    ${result1}    PTN
    ${item1}    Get From Dictionary    ${result1}    ${key}
    ${tokenResult}    transferToken    ${key}    ${GeneAdd}    ${recieverAdd}    ${senderAmount}    ${pdg}
    ...    ${evidence}    ${unlocktime}
    sleep    2
    ${item'}    Evaluate    ${item1}-${senderAmount}
    ${PTN'}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${pdg}')    decimal
    ${GeneAdd2}    getGeneAdd    ${host}
    ${result2}    getBalance    ${GeneAdd}
    ${key}    getTokenId    ${preTokenId}    ${result2}
    ${PTN2}    Get From Dictionary    ${result2}    PTN
    ${item2}    Get From Dictionary    ${result2}    ${key}
    Should Be Equal As Numbers    ${item2}    ${item'}
    Should Be Equal As Strings    ${PTN2}    ${PTN'}
