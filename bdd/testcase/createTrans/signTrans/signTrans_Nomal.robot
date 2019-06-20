*** Settings ***
Default Tags      nomal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${result_hex}     \f[a-z0-9]*
${result_txid}    \0[a-z0-9]{160,170}

*** Test Cases ***
Scenario: createTrans - Sign Transaction
    [Tags]    normal
    ${geneAdd}    getGeneAdd    ${host}
    Set Suite Variable    ${geneAdd}    ${geneAdd}
    ${ret1}    normalCrtTrans    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ${ret2}    normalSignTrans    ${ret1}    ${signType}    ${pwd}

*** Keywords ***
