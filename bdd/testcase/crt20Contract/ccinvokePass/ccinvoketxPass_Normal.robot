*** Settings ***
Suite Setup       getlistAccounts
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA050

*** Test Cases ***
CcinvokePass normal
    ${ccList}    Create List    ${crtTokenMethod}    ${evidence}    ${preTokenId}    ${tokenDecimal}    ${tokenAmount}
    ...    ${listAccounts[0]}
    normalCcinvokePass    ${commonResultCode}    ${listAccounts[0]}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}    ${20ContractId}
    ...    ${ccList}
    sleep    2
