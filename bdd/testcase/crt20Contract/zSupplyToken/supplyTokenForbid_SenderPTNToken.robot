*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA258

*** Test Cases ***
Scenario: 20Contract- Supply token
    [Documentation]    Verify Sender's PTN
    Given CcinvokePass normal
    ${PTN1}    ${key}    ${coinToken1}    And Request getbalance before create token
    ${ret}    When Create token of vote contract
    ${GAIN}    And Calculate gain of recieverAdd
    ${PTN2}    ${coinToken2}    And Request getbalance after create token    ${key}
    Then Assert gain of reciever    ${PTN1}    ${PTN2}    ${GAIN}    ${coinToken1}    ${coinToken2}    ${ret}

*** Keywords ***
CcinvokePass normal
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    ${ccList}    Create List    ${crtTokenMethod}    ${evidence}    ${preTokenId}    ${tokenDecimal}    ${tokenAmount}
    ${ret}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}

Request getbalance before create token
    sleep    4
    ${result1}    getBalance    ${geneAdd}
    ${key}    getTokenId    ${preTokenId}    ${result1}
    ${PTN1}    Get From Dictionary    ${result1}    PTN
    ${coinToken1}    Get From Dictionary    ${result1}    ${key}
    [Return]    ${PTN1}    ${key}    ${coinToken1}

Create token of vote contract
    ${ccTokenList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${supplyTokenAmount}
    ${ccList}    Create List    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}    ${20ContractId}
    ...    ${ccTokenList}    ${pwd}    ${duration}    ${EMPTY}
    ${resp}    setPostRequest    ${host}    ${invokePsMethod}    ${ccList}
    ${jsonRes}    resultToJson    ${resp}
    ${ret}    Should Match Regexp    ${jsonRes['result']}    ${commonResultCode}    msg="result:does't match Result expression"
    [Return]    ${ret}

Calculate gain of recieverAdd
    ${invokeGain}    Evaluate    int(${PTNAmount})+int(${PTNPoundage})
    ${GAIN}    countRecieverPTN    ${invokeGain}
    [Return]    ${GAIN}

Request getbalance after create token
    [Arguments]    ${key}
    sleep    4
    ${result2}    getBalance    ${geneAdd}
    ${coinToken2}    Get From Dictionary    ${result2}    ${key}
    ${PTN2}    Get From Dictionary    ${result2}    PTN
    [Return]    ${PTN2}    ${coinToken2}

Assert gain of reciever
    [Arguments]    ${PTN1}    ${PTN2}    ${GAIN}    ${coinToken1}    ${coinToken2}    ${ret}
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${GAIN}')    decimal
    Should Be Equal As Numbers    ${PTN2}    ${PTNGAIN}
    Should Be Equal As Numbers    ${coinToken1}    ${coinToken2}
    ${result}    getTxByReqId    ${ret}
    #${jsonRes}    resultToJson    ${result['result']}
    #${error_code}    Should Match Regexp    ${jsonRes['result']}    "error_code":500
    #${error_code}    Should Match Regexp    ${jsonRes['result']}    Not the supply address
    #Should Be Equal As Strings    ${jsonRes['result']['info']['contract_invoke']['error_message'] }    Chaincode Error:{\"Error\":\"Not the supply address\"}
