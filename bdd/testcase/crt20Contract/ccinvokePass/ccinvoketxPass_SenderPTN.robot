*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA052

*** Test Cases ***
Scenario: 20Contract - Create Token
    [Documentation]    Verify Sender's PTN
    ${PTN1}    Given Request getbalance before create token
    ${ret}    When Request normal CcinvokePass
    ${PTNGAIN}    And Calculate gain
    ${PTN2}    And Request getbalance after create token
    Then Assert gain    ${PTN1}    ${PTN2}    ${PTNGAIN}
    

*** Keywords ***
Request getbalance before create token
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    personalUnlockAccount    ${geneAdd}
    ${PTN1}    ${result}    normalGetBalance    ${geneAdd}
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
    ${PTN2}    ${result}    normalGetBalance    ${geneAdd}
    [Return]    ${PTN2}

Assert gain
    [Arguments]    ${PTN1}    ${PTN2}    ${PTNGAIN}
    ${GAIN}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${PTNGAIN}')    decimal
    Should Be Equal As Strings    ${PTN2}    ${GAIN}
