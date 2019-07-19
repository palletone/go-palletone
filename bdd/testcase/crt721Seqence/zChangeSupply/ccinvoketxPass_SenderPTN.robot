*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA078

*** Test Cases ***
Scenario: 721 Contract - Change token then supply token
    [Documentation]    Verify Sender's PTN and token
    Given Send the new address PTN
    ${resp1}    And CcinvokePass normal
    ${resp2}    When Supply token of 721 contract before change supply
    ${PTN2}    And Request getbalance after supply token
    ${resp3}    And Change supply address to new address
    ${PTN1}    And Request getbalance before supply token
    ${PTNGAIN}    And Calculate gain
    ${resp4}    And Supply token of 721 contract after change supply
    ${PTN3}    And Request getbalance after change supply
    Then Assert gain    ${PTN1}    ${PTN3}    ${PTNGAIN}
    #And Genesis address supply token of 721 contract
    #And Request getbalance after genesis supply token

*** Keywords ***
Send the new address PTN
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    ${jsonRes}    newAccount
    Set Suite Variable    ${reciever}    ${jsonRes['result']}
    ${ret1}    And normalCrtTrans    ${geneAdd}    ${reciever}    100000    ${PTNPoundage}
    ${ret2}    And normalSignTrans    ${ret1}    ${signType}    ${pwd}
    ${ret3}    And normalSendTrans    ${ret2}
    sleep    4

CcinvokePass normal
    ${ccList}    Create List    ${crtTokenMethod}    ${note}    ${preTokenId}    ${SeqenceToken}    ${721TokenAmount}
    ...    ${721MetaBefore}    ${geneAdd}
    ${resp1}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${reciever}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    sleep    4
    [Return]    ${resp1}

Supply token of 721 contract before change supply
    ${ccList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${721TokenAmount}    ${721MetaAfter}
    ${resp2}    normalCcinvokePass    ${commonResultCode}    ${reciever}    ${reciever}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp2}

Request getbalance after supply token
    #normalCcqueryById    ${721ContractId}    getTokenInfo    ${preTokenId}
    sleep    4
    ${PTN2}    ${result2}    normalGetBalance    ${geneAdd}
    ${key}    getTokenId    ${preTokenId}    ${result2['result']}
    log    ${key}
    ${queryResult}    ccqueryById    ${721ContractId}    ${TokenInfoMethod}    ${preTokenId}
    ${tokenCommonId}    ${countList}    jsonLoads    ${queryResult['result']}    AssetID    TokenIDs
    log    len(${countList})
    ${len}    Evaluate    len(${countList})+1
    ${queryResult}    ccqueryById    ${721ContractId}    ${TokenInfoMethod}    ${preTokenId}
    ${tokenCommonId}    ${countList}    jsonLoads    ${queryResult['result']}    AssetID    TokenIDs
    log    len(${countList})
    ${len}    Evaluate    len(${countList})
    : FOR    ${num}    IN RANGE    ${len}    1
    \    ${voteToken}    Get From Dictionary    ${result2['result']}    ${tokenCommonId}-${num}
    \    log    ${tokenCommonId}-${num}
    \    Should Be Equal As Numbers    ${voteToken}    1
    Should Not Contain    ${result2['result']}    ${tokenCommonId}-6
    [Return]    ${PTN2}

Change supply address to new address
    ${ccList}    Create List    ${changeSupplyMethod}    ${preTokenId}    ${reciever}
    ${resp3}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${geneAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp3}

Request getbalance before supply token
    sleep    4
    ${result1}    getBalance    ${reciever}
    ${PTN1}    Get From Dictionary    ${result1}    PTN
    [Return]    ${PTN1}

Calculate gain
    #${GAIN}    Evaluate    ${PTNAmount}-${PTNPoundage}
    ${PTNGAIN}    countRecieverPTN    ${PTNPoundage}
    #${PTNGAIN}    Evaluate    decimal.Decimal('${PTNAmount}')-decimal.Decimal('${PTNPoundage}')    decimal
    [Return]    ${PTNGAIN}

Supply token of 721 contract after change supply
    ${ccList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${721TokenAmount}    ${721MetaAfter}
    ${resp4}    normalCcinvokePass    ${commonResultCode}    ${reciever}    ${reciever}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp4}

Request getbalance after change supply
    sleep    4
    ${PTN3}    ${result3}    normalGetBalance    ${reciever}
    ${key}    getTokenId    ${preTokenId}    ${result3['result']}
    log    ${key}
    ${queryResult}    ccqueryById    ${721ContractId}    ${TokenInfoMethod}    ${preTokenId}
    ${tokenCommonId}    ${countList}    jsonLoads    ${queryResult['result']}    AssetID    TokenIDs
    log    len(${countList})
    ${len}    Evaluate    len(${countList})
    : FOR    ${num}    IN RANGE    6    ${len}    1
    \    ${voteToken}    Get From Dictionary    ${result3['result']}    ${tokenCommonId}-${num}
    \    log    ${tokenCommonId}-${num}
    \    Should Be Equal As Numbers    ${voteToken}    1
    [Return]    ${PTN3}

Assert gain
    [Arguments]    ${PTN1}    ${PTN3}    ${PTNGAIN}
    ${GAIN}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${PTNGAIN}')    decimal
    Should Be Equal As Strings    ${PTN3}    ${GAIN}

Genesis address supply token of 721 contract
    ${ccList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${721TokenAmount}    ${721MetaAfter}
    ${resp5}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${geneAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp5}

Request getbalance after genesis supply token
    sleep    4
    ${PTN4}    ${result4}    normalGetBalance    ${geneAdd}
    ${key}    getTokenId    ${preTokenId}    ${result4['result']}
    log    ${key}
    ${queryResult}    ccqueryById    ${721ContractId}    ${TokenInfoMethod}    ${preTokenId}
    ${tokenCommonId}    ${countList}    jsonLoads    ${queryResult['result']}    AssetID    TokenIDs
    log    len(${countList})
    ${len}    Evaluate    len(${countList})+1
    Should Not Contain    ${result4['result']}    ${tokenCommonId}-11
    [Return]    ${PTN4}
