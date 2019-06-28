*** Settings ***
Default Tags      normal
Library           ../../../utilFunc/createToken.py
Resource          ../../../utilKwd/utilVariables.txt
Resource          ../../../utilKwd/normalKwd.txt
Resource          ../../../utilKwd/utilDefined.txt
Resource          ../../../utilKwd/behaveKwd.txt

*** Variables ***

*** Test Cases ***
Scenario: Vote Contract - Create Token
    [Documentation]    Verify Sender's PTN
    ${geneAdd}    Given Get genesis address
    ${PTN1}    ${result1}    And Request getbalance before create token    ${geneAdd}
    ${ret}    When Create token of vote contract    ${geneAdd}
    ${PTNGAIN}    And Calculate gain of recieverAdd    ${PTN1}
    ${PTN}    ${result2}    And Request getbalance after create token    ${geneAdd}
    ${voteToken}    Then Assert gain of reciever    ${PTN}    ${result2}

*** Keywords ***
Get genesis address
    ${geneAdd}    getMultiNodeGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    personalUnlockAccount    ${geneAdd}
    [Return]    ${geneAdd}

Request getbalance before create token
    [Arguments]    ${geneAdd}
    sleep    4
    ${PTN1}    ${result1}    normalGetBalance    ${geneAdd}    ${mutiHost1}
    [Return]    ${PTN1}    ${result1}

Create token of vote contract
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

Calculate gain of recieverAdd
    [Arguments]    ${PTN1}
    ${totoalGain}    Evaluate    int(${amount})+int(${tokenDecimal})
    ${GAIN}    countRecieverPTN    ${totoalGain}
    [Return]    ${GAIN}

Request getbalance after create token
    [Arguments]    ${geneAdd}
    sleep    4
    ${PTN}    ${result2}    normalGetBalance    ${geneAdd}    ${mutiHost1}
    [Return]    ${PTN}    ${result2}

Assert gain of reciever
    [Arguments]    ${PTN}    ${result2}
    ${jsonRes}    Evaluate    demjson.encode(${result2})    demjson
    ${type}    Evaluate    type(${jsonRes})
    ${item}    getTokenId    ${voteId}    ${result2['result']}
    #${jsonRes}    To Json    ${jsonRes}
    #${strResult}    Evaluate    str(${jsonRes})
    ${voteToken}    Get From Dictionary    ${result2['result']}    ${item}
    Should Be Equal As Numbers    ${tokenAmount}    ${voteToken}
    [Return]    ${voteToken}
