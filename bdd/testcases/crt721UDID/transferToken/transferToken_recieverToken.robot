*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA084
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
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    personalUnlockAccount    ${geneAdd}
    sleep    3

Create token of 721 contract
    ${ccList}    Create List    ${crtTokenMethod}    ${note}    ${preTokenId}    ${UDIDToken}    ${721TokenAmount}
    ...    ${721MetaBefore}    ${geneAdd}
    ${resp}    Request CcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    ${jsonRes}    Evaluate    demjson.encode(${resp.content})    demjson
    ${jsonRes}    To Json    ${jsonRes}
    [Return]    ${jsonRes['result']}

Request getbalance before transfer token
    sleep    5
    ${PTN1}    ${result1}    normalGetBalance    ${geneAdd}
    #sleep    1
    ${queryResult}    ccqueryById    ${721ContractId}    getTokenInfo    ${preTokenId}
    ${tokenCommonId}    ${countList}    jsonLoads    ${queryResult['result']}    AssetID    TokenIDs
    ${key}    getTokenIdByNum    ${tokenCommonId}    ${result1['result']}    2
    sleep    2
    ${voteToken}    Get From Dictionary    ${result1['result']}    ${key}
    sleep    3
    [Return]    ${key}    ${voteToken}

Request transfer token
    [Arguments]    ${key}
    ${tokenResult}    transferToken    ${key}    ${geneAdd}    ${recieverAdd}    1    ${PTNPoundage}
    ...    ${evidence}    ${duration}
    sleep    5

Request getbalance after transfer token
    [Arguments]    ${key}
    ${PTN1}    ${result2}    normalGetBalance    ${recieverAdd}
    #sleep    6
    ${voteToken2}    Get From Dictionary    ${result2['result']}    ${key}
    sleep    3
    [Return]    ${voteToken2}

Assert gain
    [Arguments]    ${voteToken}    ${voteToken2}
    Should Be Equal As Strings    ${voteToken}    ${voteToken2}
