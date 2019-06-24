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
Scenario: invalidSenderAdd
    [Template]    InvalidCcinvoke
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    1    ${6000}    ${Empty}    -32000    multiple keys match address
    ...    ${EMPTY}    ${listAccounts[1]}    ${listAccounts[1]}
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    1    ${6000}    ${Empty}    -32000    multiple keys match address
    ...    P    ${listAccounts[1]}    ${listAccounts[1]}
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    1    ${6000}    ${Empty}    -32000    multiple keys match address
    ...    P1    ${listAccounts[1]}    ${listAccounts[1]}
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    1    ${6000}    ${Empty}    -32000    multiple keys match address
    ...    P1HhWxfQLMgb5TfE56GASURCuitX2XL397    ${listAccounts[1]}    ${listAccounts[1]}
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    1    ${6000}    ${Empty}    -32000    multiple keys match address
    ...    P1HhWxfQLMgb5TfE56GASURCuitX2XL397G1    ${listAccounts[1]}    ${listAccounts[1]}
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    1    ${6000}    ${Empty}    -32000    multiple keys match address
    ...    P1HhWxfQLMgb5TfE56GAFE8332432KK4JUWL    ${listAccounts[1]}    ${listAccounts[1]}

*** Keywords ***
