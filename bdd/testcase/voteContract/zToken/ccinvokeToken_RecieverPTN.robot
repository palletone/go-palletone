*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***

*** Test Cases ***
Ccinvoke Token
    [Documentation]    Scenario: Verify Reciever's PTN
    ${geneAdd}    Given Get genesis address
    ${PTN2P}    ${key}    And Request getbalance before create token    ${geneAdd}
    When Create token of vote contract    ${geneAdd}
    ${PTN2'}    And Calculate gain of recieverAdd    ${PTN2P}
    ${PTN2}    Request getbalance after create token    ${geneAdd}
    Then Assert gain of reciever    ${PTN2'}    ${PTN2}

*** Keywords ***
Get genesis address
    ${geneAdd}    getGeneAdd    ${host}
    [Return]    ${geneAdd}

Request getbalance before create token
    [Arguments]    ${geneAdd}
    ${PTN1}    ${result1}    normalGetBalance    ${geneAdd}
    ${key}    getTokenIdByNum    ${voteId}    ${result1['result']}    2
    ${PTN2}    ${result2}    normalGetBalance    ${recieverAdd}
    ${PTN2P}    voteExist    PTN    ${result2}
    [Return]    ${PTN2P}    ${key}

Create token of vote contract
    [Arguments]    ${geneAdd}
    ${supportList}    Create List    support    ${supportSection}
    ${ccList}    Create List    ${geneAdd}    ${recieverAdd}    ${destructionAdd}    ${votePTN}    ${PTNPoundage}
    ...    ${key}    ${gain}    ${voteContractId}    ${supportList}
    ${resp}    setPostRequest    ${host}    ${invokeTokenMethod}    ${ccList}
    #[Return]    ${ret}

Calculate gain of recieverAdd
    [Arguments]    ${PTN2P}
    ${GAIN}    countRecieverPTN    ${votePTN}
    ${PTN2'}    Evaluate    decimal.Decimal('${PTN2P}')+decimal.Decimal('${GAIN}')    decimal
    [Return]    ${PTN2'}

Request getbalance after create token
    [Arguments]    ${geneAdd}
    sleep    4
    ${PTN2}    ${result2}    normalGetBalance    ${recieverAdd}
    [Return]    ${PTN2}

Assert gain of reciever
    [Arguments]    ${PTN2'}    ${PTN2}
    Should Be Equal As Strings    ${PTN2}    ${PTN2'}
    ${response}    normalCcqueryById    ${voteContractId}    getVoteResult    ${key}
    ccqueryVoteResult    ${response}    ${geneAdd}    ${key}    ${tokenAmount}
