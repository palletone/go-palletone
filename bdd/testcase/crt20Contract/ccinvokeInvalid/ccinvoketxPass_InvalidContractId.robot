*** Settings ***
Suite Setup       getlistAccounts
Default Tags      invalid
Library           RequestsLibrary
Library           Collections
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/invalidKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt
Resource          ../../utilKwd/normalKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${method}         contract_ccinvoketxPass

*** Test Cases ***
Scenario: invalidContractId
    [Template]    InvalidCcinvoke
    200    2    {EMPTY}    createToken    QA666    evidence    2
    ...    1000    1    ${6000}    ${EMPTY}    -32000    ContractInvokeReq request param is error
    ...    ${listAccounts[0]}    ${listAccounts[1]}    ${listAccounts[1]}
    200    2    $    createToken    QA666    evidence    2
    ...    1000    1    ${6000}    ${EMPTY}    -32000    ContractInvokeReq request param is error
    ...    ${listAccounts[0]}    ${listAccounts[1]}    ${listAccounts[1]}
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG4    createToken    QA666    evidence    2
    ...    1000    1    ${6000}    ${EMPTY}    -32000    ContractInvokeReq request param is error
    ...    ${listAccounts[0]}    ${listAccounts[1]}    ${listAccounts[1]}
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG433    createToken    QA666    evidence    2
    ...    1000    1    ${6000}    ${EMPTY}    -32000    ContractInvokeReq request param is error
    ...    ${listAccounts[0]}    ${listAccounts[1]}    ${listAccounts[1]}
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG54    createToken    QA666    evidence    2
    ...    1000    1    ${6000}    ${EMPTY}    -32000    ContractInvokeReq request param is error
    ...    ${listAccounts[0]}    ${listAccounts[1]}    ${listAccounts[1]}

*** Keywords ***
