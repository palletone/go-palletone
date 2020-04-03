*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${preTokenId}     QA053

*** Test Cases ***
Scenario: 20Contract - Ccquery
    normalCcqueryById    ${20ContractId}    ${TokenInfoMethod}    ${preTokenId}
