*** Settings ***
Default Tags      nomal
Library           ../../../utilFunc/createToken.py
Resource          ../../../utilKwd/utilVariables.txt
Resource          ../../../utilKwd/normalKwd.txt
Resource          ../../../utilKwd/utilDefined.txt
Resource          ../../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA115

*** Test Cases ***
Scenario: 721 Contract - Supply token
    [Documentation]    Verify Sender's PTN and token
    #${ret}    Given CcinvokePass normal
    Given CcinvokePass normal
    ${PTN1}    And Request getbalance before create token
    ${ret}    When Spply token of 721 contract
    ${PTNGAIN}    Calculate gain
    ${PTN2}    Request getbalance after transfer token
    Then Assert gain    ${PTN1}    ${PTN2}    ${PTNGAIN}

*** Keywords ***
CcinvokePass normal
    ${geneAdd}    getMultiNodeGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    ${ccList}    Create List    ${crtTokenMethod}    ${note}    ${preTokenId}    ${SeqenceToken}    ${721TokenAmount}
    ...    ${721MetaBefore}    ${geneAdd}
    ${resp}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp}

Request getbalance before create token
    sleep    4
    #${PTN1}    ${result1}    normalGetBalance    ${geneAdd}
    ${result1}    getBalance    ${geneAdd}    ${mutiHost1}
    ${PTN1}    Get From Dictionary    ${result1}    PTN
    #${coinToken1}    Get From Dictionary    ${result1}    ${key}
    [Return]    ${PTN1}

Spply token of 721 contract
    ${ccList}    Create List    ${supplyTokenMethod}    ${preTokenId}    ${721TokenAmount}    ${721MetaAfter}
    ${resp}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp}

Calculate gain
    ${PTNGAIN}    Evaluate    ${PTNAmount}+${PTNPoundage}
    ${PTNGAIN}    countRecieverPTN    ${PTNGAIN}
    [Return]    ${PTNGAIN}

Request getbalance after transfer token
    sleep    4
    #normalCcqueryById    ${721ContractId}    getTokenInfo    ${preTokenId}
    ${PTN2}    ${result2}    normalGetBalance    ${geneAdd}    ${mutiHost1}
    ${key}    getTokenId    ${preTokenId}    ${result2['result']}
    log    {key}
    #${queryResult}    ccqueryById    ${721ContractId}    ${existToken}    ${key}
    #Should Be Equal As Strings    ${queryResult['result']}    True
    ${queryResult}    ccqueryById    ${721ContractId}    ${TokenInfoMethod}    ${preTokenId}
    ${tokenCommonId}    ${countList}    jsonLoads    ${queryResult['result']}    AssetID    TokenIDs
    log    len(${countList})
    ${len}    Evaluate    len(${countList})
    : FOR    ${num}    IN RANGE    6    ${len}    1
    \    ${voteToken}    Get From Dictionary    ${result2['result']}    ${tokenCommonId}-${num}
    \    log    ${tokenCommonId}-${num}
    \    Should Be Equal As Numbers    ${voteToken}    1
    [Return]    ${PTN2}

Assert gain
    [Arguments]    ${PTN1}    ${PTN2}    ${PTNGAIN}
    sleep    4
    #${result2}    getBalance    ${geneAdd}
    #${PTN2}    Get From Dictionary    ${result2}    PTN
    ${GAIN}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${PTNGAIN}')    decimal
    Should Be Equal As Numbers    ${PTN2}    ${GAIN}
