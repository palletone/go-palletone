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
Scenario: invalidContrSenderAdd
    [Template]    InvalidCcinvoke
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    ${EMPTY}    1    ${6000}    result    ContractInvokeReq request param is error
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    P    1    ${6000}    result    ContractInvokeReq request param is error
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    P1    1    ${6000}    result    ContractInvokeReq request param is error
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSx    1    ${6000}    result    ContractInvokeReq request param is error
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxwr    1    ${6000}    result    ContractInvokeReq request param is error

*** Keywords ***
