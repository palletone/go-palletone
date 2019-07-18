*** Settings ***
Library           RequestsLibrary
Library           Collections
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/invalidKwd.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${method}         wallet_getBalance
${result_code}    \f[a-z0-9]*

*** Test Cases ***
getBalance_Invalid1
    [Tags]    invalidBalance
    [Template]    invalidGetbalance
    ${empty}    -32000    Address cannot be null

getBalance_Invalid2
    [Tags]    invalidBalance
    [Template]    invalidGetbalance
    .    -32000    PalletOne address must start with 'P'
    gr#tw    -32000    PalletOne address must start with 'P'
    fdsas    -32000    PalletOne address must start with 'P'

getBalance_Invalid3
    [Tags]    invalidBalance
    [Template]    invalidGetbalance
    Pfsadf    -32000    invalid format: version and/or checksum bytes missing
