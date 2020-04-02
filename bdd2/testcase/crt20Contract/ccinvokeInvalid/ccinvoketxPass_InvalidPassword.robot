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
Scenario: invalidPassword
    [Template]    InvalidCcinvoke
    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    2    ${6000}    ${Empty}    -32000    could not decrypt key with given passphrase
    ...    ${listAccounts[0]}    ${listAccounts[1]}    ${listAccounts[1]}
    200    ${EMPTY}    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence    2
    ...    1000    2    ${6000}    ${Empty}    -32000    could not decrypt key with given passphrase
    ...    ${listAccounts[0]}    ${listAccounts[1]}    ${listAccounts[1]}

*** Keywords ***
