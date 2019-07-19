*** Settings ***
Default Tags      normal
Library           ../../../utilFunc/createToken.py
Resource          ../../../utilKwd/utilVariables.txt
Resource          ../../../utilKwd/normalKwd.txt
Resource          ../../../utilKwd/utilDefined.txt
Resource          ../../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA110

*** Test Cases ***
Feature: 721 Contract - Create token
    [Documentation]    Scenario: Verify Reciever's PTN
    ${PTN1}    ${result1}    Given Request getbalance before create token
    ${ret}    When Create token of vote contract
    ${PTNGAIN}    And Calculate gain of recieverAdd    ${PTN1}
    ${PTN2}    ${result2}    And Request getbalance after create token
    Then Assert gain of reciever    ${PTN2}    ${PTNGAIN}

*** Keywords ***
Request getbalance before create token
    ${geneAdd}    getMultiNodeGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    personalUnlockAccount    ${geneAdd}
    sleep    4
    ${PTN1}    ${result1}    normalGetBalance    ${recieverAdd}    ${mutiHost1}
    [Return]    ${PTN1}    ${result1}

Create token of vote contract
    ${ccList}    Create List    ${crtTokenMethod}    ${note}    ${preTokenId}    ${SeqenceToken}    ${721TokenAmount}
    ...    ${721MetaBefore}    ${geneAdd}
    ${resp}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp}

Calculate gain of recieverAdd
    [Arguments]    ${PTN1}
    ${gain1}    countRecieverPTN    ${PTNAmount}
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')+decimal.Decimal('${gain1}')    decimal
    [Return]    ${PTNGAIN}

Request getbalance after create token
    sleep    4
    ${PTN2}    ${result2}    normalGetBalance    ${recieverAdd}    ${mutiHost1}
    [Return]    ${PTN2}    ${result2}

Assert gain of reciever
    [Arguments]    ${PTN2}    ${PTNGAIN}
    Should Be Equal As Numbers    ${PTN2}    ${PTNGAIN}
