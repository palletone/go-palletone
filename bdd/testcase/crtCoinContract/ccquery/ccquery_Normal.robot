*** Settings ***
Default Tags      nomal
Library           RequestsLibrary
Library           Collections
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${contractId}     PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
${methodType}     getTokenInfo
${preTokenId}     QA001
${result_code}    [a-z0-9]{64}

*** Test Cases ***
ccquery_Normal
    [Documentation]    normal input
    [Tags]    normal
    normalCcqueryById    ${contractId}    ${methodType}    ${preTokenId}
