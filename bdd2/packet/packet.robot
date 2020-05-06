*** Settings ***
Library           RequestsLibrary
Library           Collections
Library           String

*** Variables ***
${tokenHolder}    ${EMPTY}
${tokenHolderPubKey}    ${EMPTY}
${signature}      ${EMPTY}

*** Test Cases ***
packet
    [Documentation]    amount = 90
    ...    count = 10
    ...    min = 1
    ...    max = 10
    listAccounts    #    主要获取 tokenHolder
    unlockAccount    ${tokenHolder}    1    #    解锁 tokenHolder
    ${twoAddr}    newAccount
    sleep    3
    transferPtn    ${tokenHolder}    ${twoAddr}    10000    1    1
    sleep    3
    unlockAccount    ${twoAddr}    1
    sleep    3
    getBalance    ${twoAddr}
    getPublicKey    ${twoAddr}
    createPacket    ${twoAddr}    90    ${tokenHolderPubKey}    10    1    10
    ...    ${EMPTY}    false
    sleep    3
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${oneAddr}    newAccount
    sign    ${twoAddr}    1
    sleep    3
    pullPacket    ${tokenHolder}    1    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    2
    sleep    3
    pullPacket    ${tokenHolder}    2    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    3
    sleep    3
    pullPacket    ${tokenHolder}    3    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    4
    sleep    3
    pullPacket    ${tokenHolder}    4    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    5
    sleep    3
    pullPacket    ${tokenHolder}    5    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    6
    sleep    3
    pullPacket    ${tokenHolder}    6    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    7
    sleep    3
    pullPacket    ${tokenHolder}    7    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    8
    sleep    3
    pullPacket    ${tokenHolder}    8    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    9
    sleep    3
    pullPacket    ${tokenHolder}    9    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    10
    sleep    3
    pullPacket    ${tokenHolder}    10    ${signature}    ${oneAddr}    0
    sleep    3
    ${amount}    getBalance    ${oneAddr}
    Should Be Equal As Numbers    ${amount}    90
    ${result}    getPacketInfo    ${tokenHolderPubKey}
    Should Be Equal As Strings    ${result["BalanceAmount"]}    0
    getPacketAllocationHistory    ${tokenHolderPubKey}
    pullPacket    ${tokenHolder}    10    ${signature}    ${oneAddr}    0
    sleep    3
    ${amount}    getBalance    ${oneAddr}
    Should Be Equal As Numbers    ${amount}    90
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${pulled}    isPulledPacket    ${tokenHolderPubKey}    10
    Should Be Equal As Strings    ${pulled}    true
    sleep    3
    sign    ${twoAddr}    11
    sleep    3
    pullPacket    ${tokenHolder}    11    ${signature}    ${oneAddr}    0
    sleep    3
    ${amount}    getBalance    ${oneAddr}
    Should Be Equal As Numbers    ${amount}    90
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}

packet1
    [Documentation]    amount = 90
    ...    count = 10
    ...    min = 1
    ...    max = 10
    ...
    ...    调整为
    ...
    ...    amount = 90
    ...    count = 11
    ...    min = 2
    ...    max = 11
    listAccounts    #    主要获取 tokenHolder
    unlockAccount    ${tokenHolder}    1    #    解锁 tokenHolder
    ${twoAddr}    newAccount
    sleep    3
    transferPtn    ${tokenHolder}    ${twoAddr}    10000    1    1
    sleep    3
    unlockAccount    ${twoAddr}    1
    getBalance    ${twoAddr}
    getPublicKey    ${twoAddr}
    createPacket    ${twoAddr}    90    ${tokenHolderPubKey}    10    1    10
    ...    ${EMPTY}    false
    sleep    3
    getPacketInfo    ${tokenHolderPubKey}
    updatePacket    ${twoAddr}    ${twoAddr}    0    ${tokenHolderPubKey}    11    2
    ...    11    ${EMPTY}    false
    sleep    3
    ${result}    getPacketInfo    ${tokenHolderPubKey}
    Should Be Equal As Strings    ${result["BalanceCount"]}    11
    getBalance    ${twoAddr}
    getAllPacketInfo

packet2
    [Documentation]    amount = 90
    ...    count = 10
    ...    min = 1
    ...    max = 10
    ...
    ...    调整为
    ...
    ...    amount = 100
    ...    count = 11
    ...    min = 2
    ...    max = 11
    listAccounts    #    主要获取 tokenHolder
    unlockAccount    ${tokenHolder}    1    #    解锁 tokenHolder
    ${twoAddr}    newAccount
    sleep    3
    transferPtn    ${tokenHolder}    ${twoAddr}    10000    1    1
    sleep    3
    unlockAccount    ${twoAddr}    1
    getBalance    ${twoAddr}
    getPublicKey    ${twoAddr}
    createPacket    ${twoAddr}    90    ${tokenHolderPubKey}    10    1    10
    ...    ${EMPTY}    false
    sleep    3
    getPacketInfo    ${tokenHolderPubKey}
    updatePacket    ${twoAddr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99    10    ${tokenHolderPubKey}    11    2
    ...    11    ${EMPTY}    TRUE
    sleep    3
    ${result}    getPacketInfo    ${tokenHolderPubKey}
    Should Be Equal As Strings    ${result["BalanceAmount"]}    100

packet3
    [Documentation]    amount = 9
    ...    count = 10
    ...    min = 1
    ...    max = 10
    listAccounts    #    主要获取 tokenHolder
    unlockAccount    ${tokenHolder}    1    #    解锁 tokenHolder
    ${twoAddr}    newAccount
    sleep    3
    transferPtn    ${tokenHolder}    ${twoAddr}    10000    1    1
    sleep    3
    unlockAccount    ${twoAddr}    1
    getBalance    ${twoAddr}
    getPublicKey    ${twoAddr}
    createPacket    ${twoAddr}    9    ${tokenHolderPubKey}    10    1    10
    ...    ${EMPTY}    false
    sleep    3
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${oneAddr}    newAccount
    sign    ${twoAddr}    1
    sleep    3
    pullPacket    ${tokenHolder}    1    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    2
    sleep    3
    pullPacket    ${tokenHolder}    2    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    3
    sleep    3
    pullPacket    ${tokenHolder}    3    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    4
    sleep    3
    pullPacket    ${tokenHolder}    4    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    5
    sleep    3
    pullPacket    ${tokenHolder}    5    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    6
    sleep    3
    pullPacket    ${tokenHolder}    6    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    7
    sleep    3
    pullPacket    ${tokenHolder}    7    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    8
    sleep    3
    pullPacket    ${tokenHolder}    8    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    9
    sleep    3
    pullPacket    ${tokenHolder}    9    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    ${result}    getPacketInfo    ${tokenHolderPubKey}
    Should Be Equal As Strings    ${result["BalanceAmount"]}    0
    Should Be Equal As Strings    ${result["BalanceCount"]}    1
    getPacketAllocationHistory    ${tokenHolderPubKey}
    pullPacket    ${tokenHolder}    9    ${signature}    ${oneAddr}    0
    sleep    3
    ${amount}    getBalance    ${oneAddr}
    Should Be Equal As Numbers    ${amount}    9
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${pulled}    isPulledPacket    ${tokenHolderPubKey}    9
    Should Be Equal As Strings    ${pulled}    true
    sleep    3
    sign    ${twoAddr}    10
    sleep    3
    pullPacket    ${tokenHolder}    10    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}

packet4
    [Documentation]    amount = 900
    ...    count = 10
    ...    min = 1
    ...    max = 10
    listAccounts    #    主要获取 tokenHolder
    unlockAccount    ${tokenHolder}    1    #    解锁 tokenHolder
    ${twoAddr}    newAccount
    sleep    3
    transferPtn    ${tokenHolder}    ${twoAddr}    10000    1    1
    sleep    3
    unlockAccount    ${twoAddr}    1
    getBalance    ${twoAddr}
    getPublicKey    ${twoAddr}
    createPacket    ${twoAddr}    900    ${tokenHolderPubKey}    10    1    10
    ...    ${EMPTY}    true
    sleep    3
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${oneAddr}    newAccount
    sleep    3
    sign    ${twoAddr}    11
    sleep    3
    pullPacket    ${tokenHolder}    1    ${signature}    ${oneAddr}    1
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    22
    sleep    3
    pullPacket    ${tokenHolder}    2    ${signature}    ${oneAddr}    2
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    33
    sleep    3
    pullPacket    ${tokenHolder}    3    ${signature}    ${oneAddr}    3
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    pullPacket    ${tokenHolder}    3    ${signature}    ${oneAddr}    3
    sleep    3
    ${amount}    getBalance    ${oneAddr}
    Should Be Equal As Numbers    ${amount}    6
    ${result}    getPacketInfo    ${tokenHolderPubKey}
    Should Be Equal As Strings    ${result["Token"][0]["BalanceAmount"]}    894
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${pulled}    isPulledPacket    ${tokenHolderPubKey}    3
    Should Be Equal As Strings    ${pulled}    true

packet5
    [Documentation]    红包过期退回
    ${time}    Get Time
    log    ${time}
    listAccounts    #    主要获取 tokenHolder
    unlockAccount    ${tokenHolder}    1    #    解锁 tokenHolder
    ${twoAddr}    newAccount
    sleep    10
    transferPtn    ${tokenHolder}    ${twoAddr}    10000    1    1
    sleep    5
    unlockAccount    ${twoAddr}    1
    sleep    5
    getPublicKey    ${twoAddr}
    createPacket    ${twoAddr}    900    ${tokenHolderPubKey}    10    1    10
    ...    ${time}    false
    sleep    10
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    getBalance    ${twoAddr}
    recyclePacket    ${twoAddr}    ${tokenHolderPubKey}
    sleep    3
    ${result}    getPacketInfo    ${tokenHolderPubKey}
    Should Be Equal As Strings    ${result["BalanceAmount"]}    0
    Should Be Equal As Strings    ${result["BalanceCount"]}    0
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${amount}    getBalance    ${twoAddr}
    Should Be Equal As Numbers    ${amount}    10000
    sign    ${twoAddr}    1
    sleep    3
    pullPacket    ${tokenHolder}    1    ${signature}    ${tokenHolder}    0

packet6
    [Documentation]    amount = 900
    ...    count = 10
    ...    min = 1
    ...    max = 10
    listAccounts    #    主要获取 tokenHolder
    unlockAccount    ${tokenHolder}    1    #    解锁 tokenHolder
    ${twoAddr}    newAccount
    sleep    3
    transferPtn    ${tokenHolder}    ${twoAddr}    10000    1    1
    sleep    3
    unlockAccount    ${twoAddr}    1
    getBalance    ${twoAddr}
    getPublicKey    ${twoAddr}
    createPacket    ${twoAddr}    900    ${tokenHolderPubKey}    10    1    10
    ...    ${EMPTY}    false
    sleep    3
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${oneAddr}    newAccount
    sleep    3
    sign    ${twoAddr}    1
    sleep    3
    pullPacket    ${tokenHolder}    1    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    2
    sleep    3
    pullPacket    ${tokenHolder}    2    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    3
    sleep    3
    pullPacket    ${tokenHolder}    3    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    4
    sleep    3
    pullPacket    ${tokenHolder}    4    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    5
    sleep    3
    pullPacket    ${tokenHolder}    5    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    6
    sleep    3
    pullPacket    ${tokenHolder}    6    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    7
    sleep    3
    pullPacket    ${tokenHolder}    7    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    8
    sleep    3
    pullPacket    ${tokenHolder}    8    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    9
    sleep    3
    pullPacket    ${tokenHolder}    9    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    10
    sleep    3
    pullPacket    ${tokenHolder}    10    ${signature}    ${oneAddr}    0
    sleep    3
    ${amount}    getBalance    ${oneAddr}
    Should Be Equal As Numbers    ${amount}    100
    ${result}    getPacketInfo    ${tokenHolderPubKey}
    Should Be Equal As Strings    ${result["BalanceAmount"]}    800
    Should Be Equal As Strings    ${result["BalanceCount"]}    0
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sleep    3
    pullPacket    ${tokenHolder}    10    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${pulled}    isPulledPacket    ${tokenHolderPubKey}    10
    Should Be Equal As Strings    ${pulled}    true
    sleep    3
    sign    ${twoAddr}    11
    sleep    3
    pullPacket    ${tokenHolder}    11    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}

packet7
    [Documentation]    amount = 30
    ...    count = 0
    ...    min = 1
    ...    max = 10
    ...
    ...    count = 0====》无限领取
    listAccounts    #    主要获取 tokenHolder
    unlockAccount    ${tokenHolder}    1    #    解锁 tokenHolder
    ${twoAddr}    newAccount
    sleep    3
    transferPtn    ${tokenHolder}    ${twoAddr}    10000    1    1
    sleep    3
    unlockAccount    ${twoAddr}    1
    sleep    3
    getBalance    ${twoAddr}
    getPublicKey    ${twoAddr}
    createPacket    ${twoAddr}    30    ${tokenHolderPubKey}    0    1    10
    ...    ${EMPTY}    false
    sleep    3
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${oneAddr}    newAccount
    sleep    3
    sign    ${twoAddr}    1
    sleep    3
    pullPacket    ${tokenHolder}    1    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    2
    sleep    3
    pullPacket    ${tokenHolder}    2    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    3
    sleep    3
    pullPacket    ${tokenHolder}    3    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    ${result}    getPacketInfo    ${tokenHolderPubKey}
    Should Be Equal As Strings    ${result["BalanceAmount"]}    0
    Should Be Equal As Strings    ${result["BalanceCount"]}    0
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    4
    sleep    3
    pullPacket    ${tokenHolder}    4    ${signature}    ${oneAddr}    0
    sleep    3
    ${amount}    getBalance    ${oneAddr}
    Should Be Equal As Numbers    ${amount}    30
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    pullPacket    ${tokenHolder}    3    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${pulled}    isPulledPacket    ${tokenHolderPubKey}    3
    Should Be Equal As Strings    ${pulled}    true
    getAllPacketInfo

packet8
    [Documentation]    amount = 30
    ...    count = 3
    ...    min = 1
    ...    max = 10
    listAccounts    #    主要获取 tokenHolder
    unlockAccount    ${tokenHolder}    1    #    解锁 tokenHolder
    sleep    3
    ${twoAddr}    newAccount
    sleep    3
    transferPtn    ${tokenHolder}    ${twoAddr}    10000    1    1
    sleep    3
    unlockAccount    ${twoAddr}    1
    sleep    3
    getBalance    ${twoAddr}
    getPublicKey    ${twoAddr}
    createPacket    ${twoAddr}    30    ${tokenHolderPubKey}    3    1    10
    ...    ${EMPTY}    false
    sleep    3
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    ${oneAddr}    newAccount
    sleep    3
    sign    ${twoAddr}    1
    sleep    3
    pullPacket    ${tokenHolder}    1    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    2
    sleep    3
    pullPacket    ${tokenHolder}    2    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    3
    sleep    3
    pullPacket    ${tokenHolder}    3    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    ${result}    getPacketInfo    ${tokenHolderPubKey}
    Should Be Equal As Strings    ${result["BalanceAmount"]}    0
    Should Be Equal As Strings    ${result["BalanceCount"]}    0
    getPacketAllocationHistory    ${tokenHolderPubKey}
    sign    ${twoAddr}    4
    sleep    3
    pullPacket    ${tokenHolder}    4    ${signature}    ${oneAddr}    0
    sleep    3
    getBalance    ${oneAddr}
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    pullPacket    ${tokenHolder}    3    ${signature}    ${oneAddr}    0
    sleep    3
    ${amount}    getBalance    ${oneAddr}
    Should Be Equal As Numbers    ${amount}    30
    getPacketInfo    ${tokenHolderPubKey}
    getPacketAllocationHistory    ${tokenHolderPubKey}
    isPulledPacket    ${tokenHolderPubKey}    3
    getAllPacketInfo

*** Keywords ***
createPacket
    [Arguments]    ${addr}    ${amount}    ${pubkey}    ${count}    ${min}    ${max}
    ...    ${expiredTime}    ${isConstant}
    ${param}    Create List    createPacket    ${pubkey}    ${count}    ${min}    ${max}
    ...    ${expiredTime}    remark    ${isConstant}
    ${two}    Create List    ${addr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99    ${amount}    1    PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99
    ...    ${param}
    ${res}    post    contract_ccinvoketx    createPacket    ${two}
    log    ${res}    #    #    Create List    createPacket    ${pubkey}
    ...    # ${count}    ${min}    ${max}    # ${expiredTime}    remark    #
    ...    # Create List    ${addr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99    PTN    ${amount}    1
    ...    # PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99    ${param}
    #    post    contract_ccinvokeToken    createPacket    ${two}
    #    ${res}
    [Return]    ${res}

post
    [Arguments]    ${method}    ${alias}    ${params}
    ${header}    Create Dictionary    Content-Type=application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=${method}    params=${params}    id=1
    Create Session    ${alias}    http://127.0.0.1:8545    #    http://127.0.0.1:8545    http://192.168.44.128:8545
    ${resp}    Post Request    ${alias}    http://127.0.0.1:8545    data=${data}    headers=${header}
    ${respJson}    To Json    ${resp.content}
    Dictionary Should Contain Key    ${respJson}    result
    ${res}    Get From Dictionary    ${respJson}    result
    [Return]    ${res}

listAccounts
    ${param}    Create List
    ${result}    post    personal_listAccounts    personal_listAccounts    ${param}
    log    ${result}
    Set Global Variable    ${tokenHolder}    ${result[0]}
    log    ${tokenHolder}

unlockAccount
    [Arguments]    ${addr}    ${pwd}
    ${param}    Create List    ${addr}    ${pwd}
    ${result}    post    personal_unlockAccount    personal_unlockAccount    ${param}
    log    ${result}
    Should Be True    ${result}

getPublicKey
    [Arguments]    ${addr}
    ${param}    Create List    ${addr}    1
    ${result}    post    personal_getPublicKey    personal_getPublicKey    ${param}
    log    ${result}
    Set Global Variable    ${tokenHolderPubKey}    ${result}
    log    ${tokenHolderPubKey}

getPacketInfo
    [Arguments]    ${tokenHolderPubKey}
    ${param}    Create List    getPacketInfo    ${tokenHolderPubKey}
    ${two}    Create List    PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99    ${param}    ${10}
    ${res}    post    contract_ccquery    getPacketInfo    ${two}
    log    ${res}
    ${addressMap}    To Json    ${res}
    [Return]    ${addressMap}

pullPacket
    [Arguments]    ${addr}    ${message}    ${signature}    ${pullAddr}    ${amount}
    ${param}    Create List    pullPacket    ${tokenHolderPubKey}    ${message}    ${signature}    ${pullAddr}
    ...    ${amount}
    ${two}    Create List    ${addr}    ${addr}    1    1    PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99
    ...    ${param}
    ${res}    post    contract_ccinvoketx    pullPacket    ${two}
    log    ${res}

getBalance
    [Arguments]    ${addr}
    ${param}    Create List    ${addr}
    ${result}    post    wallet_getBalance    wallet_getBalance    ${param}
    log    ${result}
    ${amount}    Set Variable    ${result["PTN"]}
    [Return]    ${amount}

getPacketAllocationHistory
    [Arguments]    ${pubkey}
    ${param}    Create List    getPacketAllocationHistory    ${pubkey}
    ${two}    Create List    PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99    ${param}    ${10}
    ${res}    post    contract_ccquery    getPacketAllocationHistory    ${two}
    log    ${res}

updatePacket
    [Arguments]    ${addr}    ${toaddr}    ${amount}    ${pubkey}    ${count}    ${min}
    ...    ${max}    ${expiredTime}    ${isConstant}
    ${param}    Create List    updatePacket    ${pubkey}    ${count}    ${min}    ${max}
    ...    ${expiredTime}    remark    ${isConstant}
    ${two}    Create List    ${addr}    ${toaddr}    ${amount}    1    PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99
    ...    ${param}
    ${res}    post    contract_ccinvoketx    updatePacket    ${two}
    log    ${res}
    [Return]    ${res}

recyclePacket
    [Arguments]    ${addr}    ${pubkey}
    ${param}    Create List    recyclePacket    ${pubkey}
    ${two}    Create List    ${addr}    ${addr}    0    1    PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99
    ...    ${param}
    ${res}    post    contract_ccinvoketx    recyclePacket    ${two}
    log    ${res}

newAccount
    ${param}    Create List    1
    ${result}    post    personal_newAccount    personal_newAccount    ${param}
    log    ${result}
    #    ${oneAddr}    ${result}
    [Return]    ${result}

transferPtn
    [Arguments]    ${fromAddr}    ${toAddr}    ${amount}    ${fee}    ${pwd}
    ${param}    Create List    ${fromAddr}    ${toAddr}    ${amount}    ${fee}    ${null}
    ...    ${pwd}
    ${result}    post    wallet_transferPtn    wallet_transferPtn    ${param}
    log    ${result}

sign
    [Arguments]    ${addr}    ${message}
    ${param}    Create List    ${message}    ${addr}    1
    ${result}    post    personal_sign    personal_sign    ${param}
    log    ${result}
    ${signature1} =    Get Substring    ${result}    2
    Set Global Variable    ${signature}    ${signature1}

isPulledPacket
    [Arguments]    ${tokenHolderPubKey}    ${message}
    ${param}    Create List    isPulledPacket    ${tokenHolderPubKey}    ${message}
    ${two}    Create List    PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99    ${param}    ${10}
    ${res}    post    contract_ccquery    isPulledPacket    ${two}
    log    ${res}
    [Return]    ${res}

getAllPacketInfo
    ${param}    Create List    getAllPacketInfo
    ${two}    Create List    PCGTta3M4t3yXu8uRgkKvaWd2d8DSDC6K99    ${param}    ${10}
    ${res}    post    contract_ccquery    getAllPacketInfo    ${two}
    log    ${res}
    ${addressMap}    To Json    ${res}
    log    ${addressMap}
