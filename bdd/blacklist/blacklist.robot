*** Settings ***
Library           RequestsLibrary
Library           Collections

*** Variables ***
${foundation}     ${EMPTY}
${one}            ${EMPTY}
${two}            ${EMPTY}

*** Test Cases ***
blacklist
    [Documentation]    将地址1加入黑名单，然后将地址1的 TOKEN 转给地址2
    ${o}    getBalance    ${one}    PTN
    log    ${o}
    Should Be Equal As Numbers    ${o}    10000    #Should Be Equal As Numbers
    ${t}    getBalance    ${two}    PTN
    log    ${t}
    Should Be Equal As Numbers    ${t}    10000
    ${b}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    PTN
    log    ${b}
    Should Be Equal As Numbers    ${b}    0
    ${res}    addBlacklist    ${one}    lsls
    log    ${res}
    sleep    5
    ${b}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    PTN
    log    ${b}
    Should Be Equal As Numbers    ${b}    10000
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
    ${o}    getBalance    ${one}    PTN
    log    ${o}
    Should Be Equal As Numbers    ${o}    10000
    ${t}    getBalance    ${two}    PTN
    log    ${t}
    Should Be Equal As Numbers    ${t}    20000
    ${b}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    PTN
    log    ${b}
    Should Be Equal As Numbers    ${b}    0

multiToken
    [Documentation]    先给地址2增发 token，然后将地址2加入黑名单，然后把地址2的所有 token 转给地址 P1DQA485N7r8sjUB31pKDqE2x7ZEfJxCJ2A
    ${b}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    PTN
    log    ${b}
    Should Be Equal As Numbers    ${b}    0
    ${t}    getBalance    ${two}    PTN
    log    ${t}
    Should Be Equal As Numbers    ${t}    20000
    ${result}    createToken    ${two}
    log    ${result}
    sleep    5
    ${assetId}    ccquery
    ${t}    getBalance    ${two}    PTN
    log    ${t}
    ${t}    getBalance    ${two}    ${assetId}
    log    ${t}
    Should Be Equal As Numbers    ${t}    1000
    ${res}    addBlacklist    ${two}    lsls
    log    ${res}
    sleep    5
    ${pb}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    PTN
    log    ${pb}
    Should Be Equal As Numbers    ${pb}    19999
    ${bb}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    ${assetId}
    log    ${bb}
    Should Be Equal As Numbers    ${bb}    1000
    ${res}    getBlacklistRecords
    log    ${res}
    ${res}    getBlacklistAddress
    log    ${res}
    ${res}    payout    P1DQA485N7r8sjUB31pKDqE2x7ZEfJxCJ2A    ${pb}    PTN
    log    ${res}
    sleep    5
    ${b}    getBalance    P1DQA485N7r8sjUB31pKDqE2x7ZEfJxCJ2A    PTN
    log    ${b}
    Should Be Equal As Numbers    ${b}    19999
    ${res}    payout    P1DQA485N7r8sjUB31pKDqE2x7ZEfJxCJ2A    ${bb}    ${assetId}
    log    ${res}
    sleep    5
    ${b}    getBalance    P1DQA485N7r8sjUB31pKDqE2x7ZEfJxCJ2A    ${assetId}
    log    ${b}
    Should Be Equal As Numbers    ${b}    1000
    ${b}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    PTN
    log    ${b}
    Should Be Equal As Numbers    ${b}    0
    ${b}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DRdWEXJF    ${assetId}
    log    ${b}
    Should Be Equal As Numbers    ${b}    0

*** Keywords ***
getBalance
    [Arguments]    ${address}    ${assetId}
    ${two}    Create List    ${address}
    ${result}    post    wallet_getBalance    wallet_getBalance    ${two}
    log    ${result}
    ${len}    Get Length    ${result}
    ${amount}    Set Variable     ${result["${assetId}"]}
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
    ...    ${one}
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
    ...    ${one}
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

createToken
    [Arguments]    ${address}
    ${one}    Create List    createToken    BlackListTest    black    1    1000
    ...    ${address}
    ${two}    Create List    ${address}    ${address}    0    1    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
    ...    ${one}
    ${result}    post    contract_ccinvoketx    createToken    ${two}
    [Return]    ${result}

ccquery
    ${one}    Create List    getTokenInfo    black
    ${two}    Create List    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    ${one}    ${0}
    ${result}    post    contract_ccquery    getTokenInfo    ${two}
    ${addressMap}    To Json    ${result}
    ${assetId}    Get From Dictionary    ${addressMap}    AssetID
    [Return]    ${assetId}
