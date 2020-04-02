*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA055
${result_code}    [a-z0-9]{64}

*** Test Cases ***
Scenario: 20Contract - Transfer Token
    [Documentation]    verify sender's PTN and token
    [Tags]    normal
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    personalUnlockAccount    ${geneAdd}
    ${ccList}    Create List    ${crtTokenMethod}    ${evidence}    ${preTokenId}    ${tokenDecimal}    ${tokenAmount}
    ...    ${geneAdd}
    ${ret}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    sleep    4
    ${result1}    getBalance    ${GeneAdd}
    ${key}    getTokenId    ${preTokenId}    ${result1}
    ${PTN1}    Get From Dictionary    ${result1}    PTN
    ${item1}    Get From Dictionary    ${result1}    ${key}
    ${tokenResult}    transferToken    ${key}    ${GeneAdd}    ${recieverAdd}    ${amount}    ${PTNPoundage}
    ...    ${evidence}    ${duration}
    sleep    4
    ${item'}    Evaluate    ${item1}-${amount}
    ${PTN'}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${PTNPoundage}')    decimal
    ${result2}    getBalance    ${GeneAdd}
    ${PTN2}    Get From Dictionary    ${result2}    PTN
    ${item2}    Get From Dictionary    ${result2}    ${key}
    Should Be Equal As Numbers    ${item2}    ${item'}
    Should Be Equal As Strings    ${PTN2}    ${PTN'}