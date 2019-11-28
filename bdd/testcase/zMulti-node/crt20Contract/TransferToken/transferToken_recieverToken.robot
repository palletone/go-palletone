*** Settings ***
Default Tags      normal
Library           ../../../utilFunc/createToken.py
Resource          ../../../utilKwd/utilVariables.txt
Resource          ../../../utilKwd/normalKwd.txt
Resource          ../../../utilKwd/utilDefined.txt
Resource          ../../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA104

*** Test Cases ***
Scenario: 20Contract - Transfer Token
    [Documentation]    Verify Reciever's Token
    Given Request getbalance before create token
    ${ret}    When Request normal CcinvokePass
    ${key}    ${item}    And Request getbalance after create token
    Then Assert gain    ${key}    ${item}

*** Keywords ***
Request getbalance before create token
    ${geneAdd}    getMultiNodeGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}

Request normal CcinvokePass
    ${ccList}    Create List    ${crtTokenMethod}    ${evidence}    ${preTokenId}    ${tokenDecimal}    ${tokenAmount}
    ...    ${geneAdd}
    ${ret}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    [Return]    ${ret}

Request getbalance after create token
    sleep    4
    ${PTN2}    ${result2}    normalGetBalance    ${geneAdd}    ${mutiHost1}
    ${key}    getTokenId    ${preTokenId}    ${result2['result']}
    ${item}    Set Variable    0
    ${tokenResult}    transferToken    ${key}    ${geneAdd}    ${recieverAdd}    ${gain}    ${PTNPoundage}
    ...    ${evidence}    ${duration}
    [Return]    ${key}    ${item}

Assert gain
    [Arguments]    ${key}    ${item}
    ${item1}    Evaluate    ${item}+${gain}
    sleep    4
    ${RecPTN2}    ${RecResult2}    normalGetBalance    ${recieverAdd}    ${mutiHost1}
    ${item2}    Get From Dictionary    ${RecResult2['result']}    ${key}
    Should Be Equal As Numbers    ${item2}    ${item1}
