*** Settings ***
Default Tags      invalid
Test Template     InvalidCcqueryById
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
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
ccquery_contract
    [Documentation]    invalid contractId
    ${EMPTY}    getTokenInfo    qa001    -32000    not find chainId[palletone]
    @    getTokenInfo    qa001    -32000    not find chainId[palletone]
    r    getTokenInfo    qa001    -32000    not find chainId[palletone]
    ${SPACE}    getTokenInfo    qa001    -32000    not find chainId[palletone]
    4    getTokenInfo    qa001    -32000    not find chainId[palletone]
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG4    getTokenInfo    qa001    -32000    not find chainId[palletone]
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG433    getTokenInfo    qa001    -32000    not find chainId[palletone]

ccquery_methodType
    [Documentation]    invalid methodType
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    ${EMPTY}    qa001    -32000    Unknown function
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    %    qa001    -32000    Unknown function
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    e    qa001    -32000    Unknown function
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    3    qa001    -32000    Unknown function
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    getTokenInfo1    qa001    -32000    Unknown function
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    getTokenInf    qa001    -32000    Unknown function

ccquery_preTokenId
    [Documentation]    invalid preTokenId
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    getTokenInfo    ${EMPTY}    -32000    Token not exist
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    getTokenInfo    @    -32000    Token not exist
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    getTokenInfo    e    -32000    Token not exist
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    getTokenInfo    3    -32000    Token not exist
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    getTokenInfo    getTokenInfo1    -32000    Token not exist
    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    getTokenInfo    getTokenInf    -32000    Token not exist
