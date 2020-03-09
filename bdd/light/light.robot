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
    [Documentation]    Light节点之间转账 UTXO同步 查询余额
    ...
    ...    tokenHolder转账100PTN到Light节点node_test5-》node_test5同步UTXO,并进行余额查询-》node_test5转账80PTN到Light节点node_test6-》node_test6同步UTXO,并进行余额查询-》node_test6转账50PTN到Light节点node_test7-》node_test7同步UTXO,并进行余额查询。结果：node_test7节点账户里余额有50PTN。
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
