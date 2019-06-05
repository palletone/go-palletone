*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***

*** Test Cases ***
Scenario: Vote Contract - Transfer Token
    [Documentation]    Verify Sender's PTN and VOTE value
    ${geneAdd}    Given Get genesis address
    When Transfer token of vote contract    ${geneAdd}
    ${PTN1}    ${result1}    ${item1}    ${key}    And Request getbalance before create token    ${geneAdd}
    And Request transfer token
    ${PTN'}    ${item'}    And Calculate gain of recieverAdd    ${PTN1}    ${item1}
    ${PTN2}    ${item2}    And Request getbalance after create token    ${geneAdd}    ${key}
    Then Assert gain of reciever    ${PTN'}    ${PTN2}    ${item'}    ${item2}

*** Keywords ***
Get genesis address
    ${geneAdd}    getGeneAdd    ${host}
    [Return]    ${geneAdd}

Transfer token of vote contract
    [Arguments]    ${geneAdd}
    ${ccTokenList}    Create List    ${crtTokenMethod}    ${note}    ${tokenDecimal}    ${tokenAmount}    ${voteTime}
    ...    ${commonVoteInfo}
    ${ccList}    Create List    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}    ${voteContractId}
    ...    ${ccTokenList}    ${pwd}    ${duration}    ${EMPTY}
    ${resp}    setPostRequest    ${host}    ${invokePsMethod}    ${ccList}
    log    ${resp.content}
    Should Contain    ${resp.content}['jsonrpc']    "2.0"    msg="jsonrpc:failed"
    Should Contain    ${resp.content}['id']    1    msg="id:failed"
    ${ret}    Should Match Regexp    ${resp.content}['result']    ${commonResultCode}    msg="result:does't match Result expression"
    [Return]    ${ret}

Request getbalance before create token
    [Arguments]    ${geneAdd}
    sleep    3
    ${PTN1}    ${result1}    normalGetBalance    ${geneAdd}
    sleep    4
    ${key}    getTokenId    ${voteId}    ${result1['result']}
    #${PTN1}    Get From Dictionary    ${result1['result']}    PTN
    ${item1}    Get From Dictionary    ${result1['result']}    ${key}
    [Return]    ${PTN1}    ${result1}    ${item1}    ${key}

Request transfer token
    ${tokenResult}    transferToken    ${key}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${evidence}    ${duration}

Calculate gain of recieverAdd
    [Arguments]    ${PTN1}    ${item1}
    ${item'}    Evaluate    ${item1}-${PTNAmount}
    ${PTN'}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${PTNPoundage}')    decimal
    sleep    4
    [Return]    ${PTN'}    ${item'}

Request getbalance after create token
    [Arguments]    ${geneAdd}    ${key}
    ${PTN2}    ${result2}    normalGetBalance    ${geneAdd}
    sleep    4
    #${PTN2}    Get From Dictionary    ${result2['result']}    PTN
    ${item2}    Get From Dictionary    ${result2['result']}    ${key}
    [Return]    ${PTN2}    ${item2}

Assert gain of reciever
    [Arguments]    ${PTN'}    ${PTN2}    ${item'}    ${item2}
    Should Be Equal As Strings    ${item2}    ${item'}
    Should Be Equal As Strings    ${PTN2}    ${PTN'}
