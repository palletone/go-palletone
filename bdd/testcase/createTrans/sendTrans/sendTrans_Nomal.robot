*** Settings ***
Default Tags      nomal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt
Resource          ../../utilKwd/normalKwd.txt

*** Variables ***

${PTN}            \d+
${result_code}    \f[a-z0-9]*
${result_hex}     \f[a-z0-9]*
${result_txid}    \0[a-z0-9]{60,70}
${sendResult}     [a-z0-9]*

*** Test Cases ***
Scenario: 20Contract - Send Transaction
    [Documentation]    Verify Sender's PTN
    [Tags]    normal
    ${PTN1}    Given Request getbalance before create transaction
    ${ret1}    And normalCrtTrans    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ${ret2}    And normalSignTrans    ${ret1}    ${signType}    ${pwd}
    ${ret3}    And normalSendTrans    ${ret2}
    Then Request getbalance after create transaction    ${PTN1}

*** Keywords ***
Request getbalance before create transaction
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    ${PTN1}    ${result1}    normalGetBalance    ${recieverAdd}
    sleep    5
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')+decimal.Decimal('${PTNAmount}')    decimal
    [Return]    ${PTNGAIN}

Request getbalance after create transaction
    [Arguments]    ${PTNGAIN}
    Sleep    2
    ${PTN2}    ${result2}    normalGetBalance    ${recieverAdd}
    Sleep    5
    Should Be Equal As Numbers    ${PTNGAIN}    ${PTN2}
