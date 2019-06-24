*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QZ003
${result_code}    [a-z0-9]{64}

*** Test Cases ***
Scenario: 20Contract - Frozen Token
    [Documentation]    1.create ok 2.transfer ok 3.frozen ok 4.transfer fail
    [Tags]    normal
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    personalUnlockAccount    ${geneAdd}
    sleep    2
    ${ccList}    Create List    ${crtTokenMethod}    ${evidence}    ${preTokenId}    ${tokenDecimal}    ${tokenAmount}
    ...    ${geneAdd}
    ${ret}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    sleep    5
    ${result1}    getBalance    ${GeneAdd}
    sleep    5
    ${key}    getTokenId    ${preTokenId}    ${result1}
    sleep    2
    ${PTN1}    Get From Dictionary    ${result1}    PTN
    ${item1}    Get From Dictionary    ${result1}    ${key}
    ${tokenResult}    transferToken    ${key}    ${GeneAdd}    ${recieverAdd}    ${amount}    ${PTNPoundage}
    ...    ${evidence}    ${duration}
    sleep    4
    ${item'}    Evaluate    ${item1}-${amount}
    ${PTN'}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${PTNPoundage}')    decimal
    sleep    2
    ${result2}    getBalance    ${GeneAdd}
    sleep    4
    ${PTN2}    Get From Dictionary    ${result2}    PTN
    ${item2}    Get From Dictionary    ${result2}    ${key}
    Should Be Equal As Numbers    ${item2}    ${item'}
    Should Be Equal As Strings    ${PTN2}    ${PTN'}
    sleep    1
    ${ccList}    Create List    ${frozenTokenMethod}    ${preTokenId}
    ${ret}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    sleep    4
    ${tokenInfo}    getOneTokenInfo    ${preTokenId}
    ${status}    getTokenStatus    ${tokenInfo}
    Should Not Be Equal As Numbers    ${status}    0
    sleep    1
    ${tokenResult}    transferToken    ${key}    ${GeneAdd}    ${recieverAdd}    ${amount}    ${PTNPoundage}
    ...    ${evidence}    ${duration}
    sleep    4
    ${result3}    getBalance    ${GeneAdd}
    sleep    1
    ${item3}    Get From Dictionary    ${result3}    ${key}
    Should Be Equal As Numbers    ${item3}    ${item2}
    sleep    1
    ${ccList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${amount}
    ${ret}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    sleep    4
    ${result4}    getBalance    ${GeneAdd}
    sleep    1
    ${item4}    Get From Dictionary    ${result4}    ${key}
    Should Be Equal As Numbers    ${item4}    ${item2}
    sleep    1
    ${ccList}    Create List    ${changeSupplyMethod}    ${preTokenId}    ${recieverAdd}
    ${ret}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    sleep    4
    ${tokenInfo}    getOneTokenInfo    ${preTokenId}
    ${supplyAddr}    getTokenSupplyAddr    ${tokenInfo}
    Should Be Equal As Strings    ${supplyAddr}    ${geneAdd}
    sleep    1
    