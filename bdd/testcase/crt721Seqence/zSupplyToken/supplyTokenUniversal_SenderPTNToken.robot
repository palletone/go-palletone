*** Settings ***
Default Tags      nomal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA076

*** Test Cases ***
Scenario: 721 Contract - Supply token
    [Documentation]    Verify Sender's PTN and token
    #${ret1}    Given CcinvokePass normal
    Given CcinvokePass normal
    ${PTN1}    And Request getbalance before create token
    ${ret2}    When Supply token of 721 contract
    ${PTNGAIN}    Calculate gain    ${PTN1}
    ${PTN2}    Request getbalance after transfer token
    Then Assert gain    ${PTN2}    ${PTNGAIN}

*** Keywords ***
CcinvokePass normal
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    ${jsonRes}    newAccount
    Set Suite Variable    ${reciever}    ${jsonRes['result']}
    ${ret1}    And normalCrtTrans    ${geneAdd}    ${reciever}    100000    ${PTNPoundage}
    ${ret2}    And normalSignTrans    ${ret1}    ${signType}    ${pwd}
    ${ret3}    And normalSendTrans    ${ret2}
    sleep    4
    ${ccList}    Create List    ${crtTokenMethod}    ${note}    ${preTokenId}    ${SeqenceToken}    ${721TokenAmount}
    ...    ${721MetaBefore}
    ${resp}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${reciever}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    ${jsonRes}    Evaluate    demjson.encode(${resp.content})    demjson
    ${jsonRes}    To Json    ${jsonRes}
    [Return]    ${jsonRes['result']}

Request getbalance before create token
    #${PTN1}    ${result1}    normalGetBalance    ${geneAdd}
    sleep    4
    ${result1}    getBalance    ${reciever}
    ${PTN1}    Get From Dictionary    ${result1}    PTN
    [Return]    ${PTN1}

Supply token of 721 contract
    ${ccList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${721TokenAmount}    ${721MetaAfter}
    ${resp}    Request CcinvokePass    ${commonResultCode}    ${reciever}    ${reciever}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${reps}

Calculate gain
    [Arguments]    ${PTN1}
    ${GAIN}    countRecieverPTN    ${PTNPoundage}
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${GAIN}')    decimal
    [Return]    ${PTNGAIN}

Request getbalance after transfer token
    #normalCcqueryById    ${721ContractId}    getTokenInfo    ${preTokenId}
    sleep    4
    ${PTN2}    ${result2}    normalGetBalance    ${reciever}
    ${key}    getTokenId    ${preTokenId}    ${result2['result']}
    log    ${key}
    #${queryResult}    ccqueryById    ${721ContractId}    ${existToken}    ${key}
    #Should Be Equal As Strings    ${queryResult['result']}    True
    ${queryResult}    ccqueryById    ${721ContractId}    ${TokenInfoMethod}    ${preTokenId}
    ${tokenCommonId}    ${countList}    jsonLoads    ${queryResult['result']}    AssetID    TokenIDs
    log    len(${countList})
    ${len}    Evaluate    len(${countList})+1
    : FOR    ${num}    IN RANGE    6    ${len}    1
    \    ${voteToken}    Get From Dictionary    ${result2['result']}    ${tokenCommonId}-${num}
    \    log    ${tokenCommonId}-${num}
    \    Should Be Equal As Numbers    ${voteToken}    1
    [Return]    ${PTN2}

Assert gain
    [Arguments]    ${PTN2}    ${PTNGAIN}
    #log    ${PTN1}
    #log    ${PTNGAIN}
    #${GAIN}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${721TokenAmount}')    PTNGAIN
    #${PTN2}    Evaluate    decimal.Decimal('${PTN2}')    decimal
    Should Be Equal As Strings    ${PTN2}    ${PTNGAIN}
