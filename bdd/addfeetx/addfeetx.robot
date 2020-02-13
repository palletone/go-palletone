*** Settings ***
Library           RequestsLibrary
Library           Collections
Resource          pubVariables.robot
Resource          pubFuncs.robot
Resource          setups.robot
Library           BuiltIn

*** Variables ***
${Alice}            ${EMPTY}
${Bob}              ${EMPTY}
${AliceToken}       ${EMPTY}
${BobToken}         ${EMPTY}
${foundation}       ${EMPTY}

*** Test Cases ***
addfeetx
    unlockAccount    ${foundation}
    transferPtn    ${foundation}    ${Alice}    2000
    log    ${Alice}
    log    ${AliceToken}
    transferPtn    ${foundation}    ${Bob}    2000
    sleep    1
    unlockAccount    ${Alice}
    ${AliceTokenID}=    Alice issues her personal token, amount is 100000, decimal is 1 succeed    ${Alice}    ${AliceToken}
    Set Global Variable    ${AAAliceTokenID}    ${AliceTokenID}
    sleep    5

    unlockAccount    ${Alice}
    Alice create tx withoutfee    ${AAAliceTokenID}    ${Alice}    ${Bob}    ${amount}    ${extra}    ${pwd}

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
