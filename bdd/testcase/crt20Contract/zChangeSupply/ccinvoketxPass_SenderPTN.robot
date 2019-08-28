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
Scenario: 20Contract - Change Supply Token
    [Documentation]    Verify SupplyAdd And Transfer Token
    Given Send PTN to recieverAdd
    And Request ccinvokePass and transferToken
    ${ret}    When Change supply of contract
    Then Assert the supplyAddr
    ${PTN1}    ${key}    ${coinToken1}    And Request getbalance before create token
    ${ret}    And Request supply token
    ${tokenAmount}    And Calculate gain of recieverAdd
    ${PTN2}    ${tokenGAIN}    And Request getbalance after create token    ${key}    ${tokenAmount}    ${coinToken1}
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
    sleep    4

Request ccinvokePass and transferToken
    ${ccList}    Create List    ${crtTokenMethod}    ${evidence}    ${preTokenId}    ${tokenDecimal}    ${tokenAmount}
    ...    ${geneAdd}
    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${reciever}    ${PTNAmount}    ${PTNPoundage}    ${20ContractId}
    ...    ${ccList}
    sleep    4
    ${result1}    getBalance    ${geneAdd}
    ${key}    getTokenId    ${preTokenId}    ${result1}
    ${tokenResult}    transferToken    ${key}    ${geneAdd}    ${reciever}    2000    ${PTNPoundage}
    ...    ${evidence}    ${duration}

Change supply of contract
    sleep    4
    ${ccList}    Create List    ${changeSupplyMethod}    ${preTokenId}    ${reciever}
    ${result}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${reciever}    ${PTNAmount}    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    [Return]    ${result}

Assert the supplyAddr
    sleep    4
    ${queryResult}    ccqueryById    ${20ContractId}    ${TokenInfoMethod}    ${preTokenId}
    ${SupplyAddr}    jsonLoads    ${queryResult['result']}    SupplyAddr
    Should Be Equal As Strings    ${reciever}    ${SupplyAddr}

Request getbalance before create token
    sleep    4
    ${result1}    getBalance    ${reciever}
    ${key}    getTokenId    ${preTokenId}    ${result1}
    ${PTN1}    Get From Dictionary    ${result1}    PTN
    ${coinToken1}    Get From Dictionary    ${result1}    ${key}
    [Return]    ${PTN1}    ${key}    ${coinToken1}

Request supply token
    ${ccList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${supplyTokenAmount}    ${geneAdd}
    ${ret}    normalCcinvokePass    ${commonResultCode}    ${reciever}    ${geneAdd}    10    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    [Return]    ${ret}

Calculate gain of recieverAdd
    ${invokeGain}    Evaluate    ${PTNAmount}+${PTNPoundage}
    ${tokenAmount}    countRecieverPTN    ${invokeGain}
    [Return]    ${tokenAmount}

Request getbalance after create token
    [Arguments]    ${key}    ${tokenAmount}    ${coinToken1}
    sleep    4
    ${result2}    getBalance    ${reciever}
    ${coinToken2}    Get From Dictionary    ${result2}    ${key}
    ${PTN2}    Get From Dictionary    ${result2}    PTN
    ${tokenGAIN}    Evaluate    ${coinToken2}-${coinToken1}
    [Return]    ${PTN2}    ${tokenGAIN}

Assert gain
    [Arguments]    ${PTN1}    ${PTN2}    ${tokenGAIN}    ${tokenAmount}
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${tokenAmount}')    decimal
    #${supplyTokenAmount}    Evaluate    ${supplyTokenAmount}*(10**-${tokenDecimal})
    Should Be Equal As Numbers    ${supplyTokenAmount}    ${tokenGAIN}
