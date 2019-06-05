*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA056

*** Test Cases ***
Scenario: Change Supply Token
    [Documentation]    Verify SupplyAdd And Transfer Token
    Given Send PTN to recieverAdd
    And Request ccinvokePass and transferToken
    ${ret}    When Change supply of contract
    Then Assert the supplyAddr
    ${PTN1}    ${key}    ${coinToken1}    And Request getbalance before create token
    ${ret}    And Request supply token
    ${tokenAmount}    And Calculate gain of recieverAdd
    ${PTN2}    ${tokenGAIN}    And Request getbalance after create token    ${geneAdd}    ${key}    ${tokenAmount}
    Then Assert gain    ${PTN1}    ${PTN2}    ${tokenGAIN}    ${tokenAmount}

*** Keywords ***
Send PTN to recieverAdd
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    ${jsonRes}    newAccount
    Set Suite Variable    ${reciever}    ${jsonRes['result']}
    ${ret1}    And normalCrtTrans    ${geneAdd}    ${reciever}    100000    ${PTNPoundage}
    ${ret2}    And normalSignTrans    ${ret1}    ${signType}    ${pwd}
    ${ret3}    And normalSendTrans    ${ret2}

Request ccinvokePass and transferToken
    ${ccList}    Create List    ${crtTokenMethod}    ${evidence}    ${preTokenId}    ${tokenDecimal}    ${tokenAmount}
    ...    ${geneAdd}
    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${reciever}    ${PTNAmount}    ${PTNPoundage}    ${20ContractId}
    ...    ${ccList}
    sleep    3
    ${result1}    getBalance    ${geneAdd}
    sleep    5
    ${key}    getTokenId    ${preTokenId}    ${result1}
    sleep    2
    ${tokenResult}    transferToken    ${key}    ${geneAdd}    ${reciever}    2000    ${PTNPoundage}
    ...    ${evidence}    ${duration}

Change supply of contract
    ${ccList}    Create List    ${changeSupplyMethod}    ${preTokenId}    ${reciever}
    ${result}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${reciever}    ${PTNAmount}    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    sleep    2
    [Return]    ${result}

Assert the supplyAddr
    ${queryResult}    ccqueryById    ${20ContractId}    ${TokenInfoMethod}    ${preTokenId}
    ${SupplyAddr}    jsonLoads    ${queryResult['result']}    SupplyAddr
    Should Be Equal As Strings    ${reciever}    ${SupplyAddr}

Request getbalance before create token
    ${result1}    getBalance    ${reciever}
    sleep    5
    ${key}    getTokenId    ${preTokenId}    ${result1}
    sleep    1
    ${PTN1}    Get From Dictionary    ${result1}    PTN
    sleep    1
    ${coinToken1}    Get From Dictionary    ${result1}    ${key}
    [Return]    ${PTN1}    ${key}    ${coinToken1}

Request supply token
    ${ccList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${supplyTokenAmount}    ${geneAdd}
    ${ret}    normalCcinvokePass    ${commonResultCode}    ${reciever}    ${geneAdd}    10    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    [Return]    ${ret}

Calculate gain of recieverAdd
    ${invokeGain}    Evaluate    int(${PTNAmount})+int(${PTNPoundage})
    ${tokenAmount}    countRecieverPTN    ${invokeGain}
    sleep    3
    [Return]    ${tokenAmount}

Request getbalance after create token
    [Arguments]    ${geneAdd}    ${key}    ${tokenAmount}
    ${result2}    getBalance    ${reciever}
    sleep    5
    ${coinToken2}    Get From Dictionary    ${result2}    ${key}
    sleep    1
    ${PTN2}    Get From Dictionary    ${result2}    PTN
    sleep    1
    ${tokenGAIN}    Evaluate    float(${coinToken2})-float(${coinToken1})
    [Return]    ${PTN2}    ${tokenGAIN}

Assert gain
    [Arguments]    ${PTN1}    ${PTN2}    ${tokenGAIN}    ${tokenAmount}
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${tokenAmount}')    decimal
    ${supplyTokenAmount}    Evaluate    ${supplyTokenAmount}*(10**-${tokenDecimal})
    sleep    1
    Should Be Equal As Numbers    ${supplyTokenAmount}    ${tokenGAIN}
