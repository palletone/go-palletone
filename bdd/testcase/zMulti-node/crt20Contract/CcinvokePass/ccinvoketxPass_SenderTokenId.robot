*** Settings ***
Default Tags      normal
Library           ../../../utilFunc/createToken.py
Resource          ../../../utilKwd/utilVariables.txt
Resource          ../../../utilKwd/normalKwd.txt
Resource          ../../../utilKwd/utilDefined.txt
Resource          ../../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA103

*** Test Cases ***
Scenario: 20Contract - Create Token
    [Documentation]    Verify Sender's Token
    ${PTN1}    Given Request getbalance before create token
    ${ret}    When Request normal CcinvokePass
    ${PTNGAIN}    And Calculate gain
    ${count}    ${key}    And Request getbalance after create token
    Then Assert gain    ${count}    ${key}

*** Keywords ***
Request getbalance before create token
    ${geneAdd}    getMultiNodeGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    personalUnlockAccount    ${geneAdd}
    ${PTN1}    ${result}    normalGetBalance    ${geneAdd}    ${mutiHost1}
    [Return]    ${PTN1}

Request normal CcinvokePass
    ${ccList}    Create List    ${crtTokenMethod}    ${evidence}    ${preTokenId}    ${tokenDecimal}    ${tokenAmount}
    ...    ${geneAdd}
    ${ret}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    [Return]    ${ret}

Calculate gain
    ${PTNGAIN}    Evaluate    ${PTNAmount}+${PTNPoundage}
    ${PTNGAIN}    countRecieverPTN    ${PTNGAIN}
    [Return]    ${PTNGAIN}

Request getbalance after create token
    sleep    4
    ${PTN2}    ${result2}    normalGetBalance    ${geneAdd}    ${mutiHost1}
    : FOR    ${key}    IN    ${result2.keys}
    \    log    ${key}
    #${count}    evaluate    int(pow(10,-${tokenDecimal})*${tokenAmount})
    #log    ${result2['result']}
    ${item}    getTokenId    ${preTokenId}    ${result2['result']}
    ${key}    Get From Dictionary    ${result2['result']}    ${item}
    [Return]    ${tokenAmount}    ${key}

Assert gain
    [Arguments]    ${count}    ${key}
    Should Be Equal As Numbers    ${count}    ${key}
