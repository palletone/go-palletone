*** Settings ***
Suite Setup       voteTransToken
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${votePTN}        1000

*** Test Cases ***
Scenario: Vote - Ccinvoke Token
    [Documentation]    Verify Sender's PTN and VOTE value
    #${geneAdd}    Given Get genesis address
    ${PTN1}    ${item1}    And Request getbalance before create token
    When Ccinvoke token of vote contract
    ${PTN'}    ${item'}    And Calculate gain of recieverAdd    ${PTN1}    ${item1}
    ${PTN2}    ${item2}    Request getbalance after create token
    Then Assert gain of reciever    ${PTN'}    ${PTN2}    ${item'}    ${item2}

*** Keywords ***
Get genesis address
    ${geneAdd}    getGeneAdd    ${host}
    [Return]    ${geneAdd}

Request getbalance before create token
    #    [Arguments]    ${geneAdd}    ${voteToken}
    ${PTN1}    ${result1}    normalGetBalance    ${listAccounts[0]}
    #${key}    getTokenId    ${voteId}    ${result1['result']}
    ${item1}    Get From Dictionary    ${result1['result']}    ${voteToken}
    [Return]    ${PTN1}    ${item1}

Ccinvoke token of vote contract
    #[Arguments]    ${geneAdd}
    ${supportList}    Create List    support    ${supportSection}
    ${ccList}    Create List    ${listAccounts[0]}    ${recieverAdd}    ${destructionAdd}    ${votePTN}    ${PTNPoundage}
    ...    ${voteToken}    ${voteAmount}    ${voteContractId}    ${supportList}
    ${resp}    setPostRequest    ${host}    ${invokeTokenMethod}    ${ccList}
    log    ${resp.content}

Calculate gain of recieverAdd
    [Arguments]    ${PTN1}    ${item1}
    ${item'}    Evaluate    ${item1}-${voteAmount}
    ${totalGain}    Evaluate    int(${PTNPoundage})+int(${votePTN})
    ${GAIN}    countRecieverPTN    ${totalGain}
    ${PTN'}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${GAIN}')    decimal
    [Return]    ${PTN'}    ${item'}

Request getbalance after create token
    #[Arguments]    ${geneAdd}    ${voteToken}
    sleep    4
    ${PTN2}    ${result2}    normalGetBalance    ${listAccounts[0]}
    ${item2}    Get From Dictionary    ${result2['result']}    ${voteToken}
    [Return]    ${PTN2}    ${item2}

Assert gain of reciever
    [Arguments]    ${PTN'}    ${PTN2}    ${item'}    ${item2}
    Should Be Equal As Strings    ${item2}    ${item'}
    Should Be Equal As Numbers    ${PTN2}    ${PTN'}
