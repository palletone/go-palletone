*** Settings ***
Library           RequestsLibrary
Library           Collections

*** Variables ***
${moneyaccount}    ${EMPTY}
${account5}       ${EMPTY}
${account6}       ${EMPTY}
${account7}       ${EMPTY}

*** Test Cases ***
ligth
    transferPTN    ${moneyaccount}    ${account5}    100    1    1    http://127.0.0.1:8545
    sleep    5
    ${result}    syncUTXOByAddr    ${account5}    http://127.0.0.1:8595
    log    ${result}
    Should Be Equal As Strings    ${result}    ok
    sleep    5
    ${result}    getBalance    ${account5}    http://127.0.0.1:8595
    log    ${result}
    Should Be Equal As Integers    ${result}    100
    sleep    5
    transferPTN    ${account5}    ${account6}    80    1    1    http://127.0.0.1:8595
    sleep    5
    ${result}    syncUTXOByAddr    ${account6}    http://127.0.0.1:8605
    log    ${result}
    Should Be Equal As Strings    ${result}    ok
    sleep    5
    ${result}    getBalance    ${account6}    http://127.0.0.1:8605
    log    ${result}
    Should Be Equal As Integers    ${result}    80
    sleep    5
    transferPTN    ${account6}    ${account7}    80    1    1    http://127.0.0.1:8605
    sleep    5
    ${result}    syncUTXOByAddr    ${account7}    http://127.0.0.1:8615
    log    ${result}
    Should Be Equal As Strings    ${result}    ok
    sleep    5
    ${result}    getBalance    ${account7}    http://127.0.0.1:8615
    log    ${result}
    Should Be Equal As Integers    ${result}    50

*** Keywords ***
transferPtn
    [Arguments]    ${fromAddr}    ${toAddr}    ${amount}    ${fee}    ${pwd}    ${hostPort}
    ${param}    Create List    ${fromAddr}    ${toAddr}    ${amount}    ${fee}    ${null}
    ...    ${pwd}
    ${result}    post    wallet_transferPtn    wallet_transferPtn    ${param}    ${hostPort}
    log    ${result}

post
    [Arguments]    ${method}    ${alias}    ${params}    ${hostPort}
    ${header}    Create Dictionary    Content-Type=application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=${method}    params=${params}    id=1
    Create Session    ${alias}    ${hostPort}
    ${resp}    Post Request    ${alias}    ${hostPort}    data=${data}    headers=${header}
    ${respJson}    To Json    ${resp.content}
    Dictionary Should Contain Key    ${respJson}    result
    ${res}    Get From Dictionary    ${respJson}    result
    [Return]    ${res}

syncUTXOByAddr
    [Arguments]    ${addr}    ${hostPort}
    ${param}    Create List    ${addr}
    ${result}    post    ptn_syncUTXOByAddr    ptn_syncUTXOByAddr    ${param}    ${hostPort}
    log    ${result}
    [Return]    ${result}

getBalance
    [Arguments]    ${addr}    ${hostPort}
    ${param}    Create List    ${addr}
    ${result}    post    wallet_getBalance    wallet_getBalance    ${param}    ${hostPort}
    log    ${result}
    ${amount}    Set Variable    ${result["PTN"]}
    [Return]    ${amount}
