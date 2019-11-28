*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     CA081

*** Test Cases ***
Feature: 721 Contract - Create token
    [Documentation]    Scenario: Verify Sender's PTN
    Given Get genesis address
    ${PTN1}    ${result1}    And Request getbalance before create token
    ${ret}    When Create token of vote contract
    ${GAIN}    And Calculate gain of recieverAdd    ${PTN1}
    ${PTN2}    ${result2}    And Request getbalance after create token
    Then Assert gain of reciever    ${PTN1}    ${PTN2}    ${GAIN}

*** Keywords ***
Get genesis address
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    personalUnlockAccount    ${geneAdd}

Request getbalance before create token
    ${PTN1}    ${result1}    normalGetBalance    ${geneAdd}
    [Return]    ${PTN1}    ${result1}

Create token of vote contract
    ${ccList}    Create List    ${crtTokenMethod}    ${note}    ${preTokenId}    ${UDIDToken}    ${721TokenAmount}
    ...    ${721MetaBefore}    ${geneAdd}
    ${resp}    normalCcinvokePass    ${commonResultCode}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${721ContractId}    ${ccList}
    [Return]    ${resp}

Calculate gain of recieverAdd
    [Arguments]    ${PTN1}
    ${invokeGain}    Evaluate    int(${PTNAmount})+int(${PTNPoundage})
    ${GAIN}    countRecieverPTN    ${invokeGain}
    [Return]    ${GAIN}

Request getbalance after create token
    sleep    4
    ${PTN2}    ${result2}    normalGetBalance    ${geneAdd}
    [Return]    ${PTN2}    ${result2}

Assert gain of reciever
    [Arguments]    ${PTN1}    ${PTN2}    ${GAIN}
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${GAIN}')    decimal
    Should Be Equal As Numbers    ${PTN2}    ${PTNGAIN}
