*** Settings ***
Default Tags      nomal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
#${reciever}      P18GoFijgVSekVLF8oSzNijTKqqsorWV9fg

*** Test Cases ***
Scenario: Multi-node Create Transaction
    [Tags]    normal
    ${PTNGAIN}    Given Get multi-node genesis address
    ${ret1}    And normalCrtTrans    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ${ret2}    And normalSignTrans    ${ret1}    ${signType}    ${pwd}
    ${ret3}    And normalSendTrans    ${ret2}
    Then Assert PTN from another node    ${PTNGAIN}

*** Keywords ***
Get multi-node genesis address
    ${geneAdd}    getMultiNodeGeneAdd    ${host}
    #${reciever}    newAccount
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    #Set Suite Variable    ${reciever}    ${reciever['result']}
    #${foundationAdd1}    getGeneAdd    ${mutiHost1}
    sleep    4
    ${PTN1}    ${reuslt1}    normalGetBalance    ${recieverAdd}    ${mutiHost1}
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')+decimal.Decimal('${PTNAmount}')    decimal
    [Return]    ${PTNGAIN}

Assert PTN from another node
    [Arguments]    ${PTNGAIN}
    #${foundationAdd2}    getGeneAdd    ${mutiHost1}
    sleep    4
    ${PTN2}    ${reuslt2}    normalGetBalance    ${recieverAdd}    ${mutiHost1}
    Should Be Equal As Numbers    ${PTNGAIN}    ${PTN2}
