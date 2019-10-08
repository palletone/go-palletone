*** Settings ***
Library           RequestsLibrary
Library           Collections

*** Variables ***
${foundation}     ${EMPTY}
${one}            ${EMPTY}
${two}            ${EMPTY}

*** Test Cases ***
blacklist
    ${f}    getBalance    ${foundation}    PTN
    log    ${f}
    Should Be Equal As Numbers    ${f}    999940044
    ${o}    getBalance    ${one}    PTN
    log    ${o}
    Should Be Equal As Numbers    ${o}    9989
    ${t}    getBalance    ${two}    PTN
    log    ${t}
    Should Be Equal As Numbers    ${t}    9998
    ${b}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    PTN
    log    ${b}
    Should Be Equal As Numbers    ${b}    0
    ${res}    addBlacklist    ${one}    lsls
    log    ${res}
    sleep    5
    ${b}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    PTN
    log    ${b}
    Should Be Equal As Numbers    ${b}    9989
    ${res}    getBlacklistRecords
    log    ${res}
    ${res}    getBlacklistAddress
    log    ${res}
    Dictionary Should Contain Key    ${res}    ${one}
    ${res}    transferPTN    ${one}    ${two}    10
    log    ${res}
    Should Be Equal As Strings    ${res}    ADDRESS_IN_BLACKLIST
    ${res}    transferPTN    ${two}    ${one}    10
    log    ${res}
    Should Be Equal As Strings    ${res}    ADDRESS_IN_BLACKLIST
    ${res}    payout    ${two}    ${o}    PTN
    log    ${res}
    sleep    5
    ${f}    getBalance    ${foundation}    PTN
    log    ${f}
    Should Be Equal As Numbers    ${f}    999940042
    ${o}    getBalance    ${one}    PTN
    log    ${o}
    Should Be Equal As Numbers    ${o}    9989
    ${t}    getBalance    ${two}    PTN
    log    ${t}
    Should Be Equal As Numbers    ${t}    19987
    ${b}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    PTN
    log    ${b}
    Should Be Equal As Numbers    ${b}    0

*** Keywords ***
getBalance
    [Arguments]    ${address}    ${assetId}
    ${two}    Create List    ${address}
    ${result}    post    wallet_getBalance    wallet_getBalance    ${two}
    log    ${result}
    ${len}    Get Length    ${result}
    ${amount}    Set Variable If    ${len}==0    0    ${result["${assetId}"]}
    [Return]    ${amount}

post
    [Arguments]    ${method}    ${alias}    ${params}
    ${header}    Create Dictionary    Content-Type=application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=${method}    params=${params}    id=1
    Create Session    ${alias}    http://127.0.0.1:8545
    ${resp}    Post Request    ${alias}    http://127.0.0.1:8545    data=${data}    headers=${header}
    ${respJson}    To Json    ${resp.content}
    Dictionary Should Contain Key    ${respJson}    result
    ${res}    Get From Dictionary    ${respJson}    result
    [Return]    ${res}

addBlacklist
    [Arguments]    ${address}    ${reason}
    ${one}    Create List    addBlacklist    ${address}    ${reason}
    ${two}    Create List    ${foundation}    ${foundation}    1    1    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF
    ...    ${one}    \    10
    ${res}    post    contract_ccinvoketx    addBlacklist    ${two}
    [Return]    ${res}

getBlacklistRecords
    ${one}    Create List    getBlacklistRecords
    ${two}    Create List    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    ${one}    ${10}
    ${res}    post    contract_ccquery    getBlacklistRecords    ${two}
    [Return]    ${res}

getBlacklistAddress
    ${one}    Create List    getBlacklistAddress
    ${two}    Create List    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    ${one}    ${10}
    ${res}    post    contract_ccquery    getBlacklistAddress    ${two}
    ${addressMap}    To Json    ${res}
    [Return]    ${addressMap}

payout
    [Arguments]    ${address}    ${amount}    ${assetId}
    ${one}    Create List    payout    ${address}    ${amount}    ${assetId}
    ${two}    Create List    ${foundation}    ${foundation}    1    1    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF
    ...    ${one}    \    10
    ${res}    post    contract_ccinvoketx    payout    ${two}
    [Return]    ${res}

transferPTN
    [Arguments]    ${from}    ${to}    ${amount}
    ${one}    Create List    ${from}    ${to}    ${amount}    1    "extra"
    ...    1    ${10}
    ${header}    Create Dictionary    Content-Type=application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=wallet_transferPtn    params=${one}    id=1
    Create Session    transferPTN    http://127.0.0.1:8545
    ${resp}    Post Request    transferPTN    http://127.0.0.1:8545    data=${data}    headers=${header}
    ${respJson}    To Json    ${resp.content}
    Dictionary Should Contain Key    ${respJson}    error
    ${res}    Get From Dictionary    ${respJson}    error
    Dictionary Should Contain Key    ${res}    message
    ${res}    Get From Dictionary    ${res}    message
    [Return]    ${res}
