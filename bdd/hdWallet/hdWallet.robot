*** Settings ***
Library           RequestsLibrary
Library           Collections

*** Variables ***
${seedAddress}    ${EMPTY}
${seedPasswd}     123456
${seedMnemonic}    ${EMPTY}
${tokenHolder}    ${EMPTY}

*** Test Cases ***
hdWallet
    listAccounts    #    主要获取 tokenHolder
    sleep    3
    unlockAccount    ${tokenHolder}    1    #    解锁 tokenHolder
    sleep    3
    newHdAccount    #    创建 HD 种子
    sleep    3
    unlockAccount    ${seedAddress}    ${seedPasswd}    #    解锁 HD 种子
    sleep    3
    getHdAccount    1    #    获得HD钱包中某AccountIndex对应的地址    AccountIndex = 1
    sleep    3
    transferPTN    ${tokenHolder}    ${seedAddress}:1    10000    1    1    #
    ...    # 从 tokenHolder 转账某个HD钱包的AccountIndex
    sleep    3
    ${addr1}    getHdAccount    1    #    获得HD钱包中某AccountIndex对应的地址    AccountIndex = 1
    ${amount}    getBalance    ${addr1}    #    获取账户余额
    Should Be Equal As Integers    ${amount}    10000
    sleep    3
    getHdAccount    2    #    获得HD钱包中某AccountIndex对应的地址    AccountIndex = 2
    sleep    3
    transferPTN    ${seedAddress}:1    ${seedAddress}:2    10    1    ${seedPasswd}    #
    ...    # 从某个HD钱包的AccountIndex = 1转移到AccountIndex = 2
    sleep    3
    ${addr2}    getHdAccount    2    #    获得HD钱包中某AccountIndex对应的地址    AccountIndex = 2
    ${amount}    getBalance    ${addr2}    #    获取账户余额
    Should Be Equal As Integers    ${amount}    10

*** Keywords ***
newHdAccount
    ${param}    Create List    ${seedPasswd}
    ${result}    post    personal_newHdAccount    personal_newHdAccount    ${param}
    log    ${result}
    log    ${result["Address"]}
    log    ${result["Mnemonic"]}
    Set Global Variable    ${seedAddress}    ${result["Address"]}
    Set Global Variable    ${seedMnemonic}    ${result["Mnemonic"]}

getHdAccount
    [Arguments]    ${userId}
    ${param}    Create List    ${seedAddress}    ${seedPasswd}    ${userId}
    ${result}    post    personal_getHdAccount    personal_getHdAccount    ${param}
    log    ${result}
    [Return]    ${result}

transferPtn
    [Arguments]    ${fromAddr}    ${toAddr}    ${amount}    ${fee}    ${pwd}
    ${param}    Create List    ${fromAddr}    ${toAddr}    ${amount}    ${fee}    ${null}
    ...    ${pwd}
    ${result}    post    wallet_transferPtn    wallet_transferPtn    ${param}
    log    ${result}

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

importHdAccountMnemonic
    ${param}    Create List    ${seedMnemonic}    ${seedPasswd}
    ${result}    post    personal_importHdAccountMnemonic    personal_importHdAccountMnemonic    ${param}
    log    ${result}

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

getBalance
    [Arguments]    ${addr}
    ${param}    Create List    ${addr}
    ${result}    post    wallet_getBalance    wallet_getBalance    ${param}
    log    ${result}
    ${amount}    Set Variable    ${result["PTN"]}
    [Return]    ${amount}
