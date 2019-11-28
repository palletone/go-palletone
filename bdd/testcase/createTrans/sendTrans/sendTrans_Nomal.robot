*** Settings ***
Default Tags      nomal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt
Resource          ../../utilKwd/normalKwd.txt

*** Variables ***

*** Test Cases ***
Scenario: createTrans - Send Transaction
    [Documentation]    Verify PTN after sending
    [Tags]    normal
    sleep    4
    ${PTN1}    Given Request getbalance before create transaction
    ${ret1}    And normalCrtTrans    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ${ret2}    And normalSignTrans    ${ret1}    ${signType}    ${pwd}
    ${ret3}    And normalSendTrans    ${ret2}
    sleep    4
    Then Request getbalance after create transaction    ${PTN1}

*** Keywords ***
Request getbalance before create transaction
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    ${PTN1}    ${result1}    normalGetBalance    ${recieverAdd}
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')+decimal.Decimal('${PTNAmount}')    decimal
    [Return]    ${PTNGAIN}

Request getbalance after create transaction
    [Arguments]    ${PTNGAIN}
    Sleep    4
    ${PTN2}    ${result2}    normalGetBalance    ${recieverAdd}
    Should Be Equal As Numbers    ${PTNGAIN}    ${PTN2}
