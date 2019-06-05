*** Settings ***
Suite Setup       getlistAccounts
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA051
#${gain}          2000

*** Test Cases ***
Scenario: 20Contract - Create Token
    [Documentation]    Verify Sender's PTN
    ${PTN1}    Given Request getbalance before create token
    ${ret}    When Request normal CcinvokePass
    ${PTNGAIN}    And Calculate gain    ${PTN1}
    ${PTN2}    And Request getbalance after create token
    Then Assert gain    ${PTN2}    ${PTNGAIN}

*** Keywords ***
Request getbalance before create token
    ${PTN1}    ${result}    normalGetBalance    ${recieverAdd}
    sleep    1
    [Return]    ${PTN1}

Request normal CcinvokePass
    ${ccList}    Create List    ${crtTokenMethod}    ${evidence}    ${preTokenId}    ${tokenDecimal}    ${tokenAmount}
    ...    ${listAccounts[0]}
    ${ret}    normalCcinvokePass    ${commonResultCode}    ${listAccounts[0]}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${20ContractId}    ${ccList}
    sleep    3
    [Return]    ${ret}

Calculate gain
    [Arguments]    ${PTN1}
    ${gain1}    countRecieverPTN    ${PTNAmount}
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')+decimal.Decimal('${gain1}')    decimal
    sleep    2
    [Return]    ${PTNGAIN}

Request getbalance after create token
    ${PTN2}    ${result}    normalGetBalance    ${recieverAdd}
    sleep    5
    [Return]    ${PTN2}

Assert gain
    [Arguments]    ${PTN2}    ${PTNGAIN}
    Should Be Equal As Numbers    ${PTN2}    ${PTNGAIN}
