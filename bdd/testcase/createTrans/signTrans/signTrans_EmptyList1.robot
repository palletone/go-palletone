*** Settings ***
Force Tags        invalidSign
Default Tags      invalidSign
Library           RequestsLibrary
Library           Collections
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/invalidKwd.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${method}         ptn_signRawTransaction
${error_code}     -32000
${error_message}    Params is empty

*** Test Cases ***
signTransInvalid1
    [Tags]    invalidSign1
    [Template]    setEmptySign
    ${EMPTY}
    2

*** Keywords ***
