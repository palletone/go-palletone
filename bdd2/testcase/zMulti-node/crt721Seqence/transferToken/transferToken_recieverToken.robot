*** Settings ***
Default Tags      normal
Library           ../../../utilFunc/createToken.py
Resource          ../../../utilKwd/utilVariables.txt
Resource          ../../../utilKwd/normalKwd.txt
Resource          ../../../utilKwd/utilDefined.txt
Resource          ../../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA113
${subId}          4

*** Test Cases ***
Scenario: 721 Contract - Transfer token
    [Documentation]    Verify Reciever's Token
    Given Get genesis address
    ${ret}    When Create token of 721 contract
    ${key}    ${voteToken}    And Request getbalance before transfer token
    And Request transfer token    ${key}
    ${voteToken2}    And Request getbalance after transfer token    ${key}
    Then Assert gain    ${voteToken}    ${voteToken2}

*** Keywords ***
Get genesis address
    ${geneAdd}    getMultiNodeGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    personalUnlockAccount    ${geneAdd}

Create token of 721 contract
    ${ccList}    Create List    ${crtTokenMethod}    ${note}    ${preTokenId}    ${SeqenceToken}    ${721TokenAmount}
    ...    ${721MetaBefore}    ${geneAdd}
    ${resp}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp}

Request getbalance before transfer token
    sleep    4
    ${PTN1}    ${result1}    normalGetBalance    ${geneAdd}    ${mutiHost1}
    ${queryResult}    ccqueryById    ${721ContractId}    getTokenInfo    ${preTokenId}
    ${tokenCommonId}    ${countList}    jsonLoads    ${queryResult['result']}    AssetID    TokenIDs
    Set Suite Variable    ${key}    ${tokenCommonId}-${subId}
    ${voteToken}    Get From Dictionary    ${result1['result']}    ${key}
    [Return]    ${key}    ${voteToken}

Request transfer token
    [Arguments]    ${key}
    ${tokenResult}    transferToken    ${key}    ${geneAdd}    ${recieverAdd}    1    ${PTNPoundage}
    ...    ${evidence}    ${duration}

Request getbalance after transfer token
    [Arguments]    ${key}
    sleep    4
    ${PTN1}    ${result2}    normalGetBalance    ${recieverAdd}    ${mutiHost1}
    ${voteToken2}    Get From Dictionary    ${result2['result']}    ${key}
    [Return]    ${voteToken2}

Assert gain
    [Arguments]    ${voteToken}    ${voteToken2}
    Should Be Equal As Strings    ${voteToken}    ${voteToken2}
