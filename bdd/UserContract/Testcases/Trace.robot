*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot

*** Test Cases ***
install
    Given Unlock token holder succeed
    ${reqId} =    When User installs contract template    github.com/palletone/go-palletone/contracts/example/go/trace    trace
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

deploy
    Given Unlock token holder succeed
    ${reqId} =    When User deploys contract
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    sleep    10

invoke01
    [Documentation]    tokenholder
    GetAdmin
    Given Unlock token holder succeed
    ${reqId} =    AddProof    ${tokenHolder}    enterprises    enterprisesId1    enterprisesMessage1    enterprisesId1
    ...    ${tokenHolder}
    sleep    10
    GetTxByReqId    ${reqId}
    ${reqId} =    AddProof    ${tokenHolder}    goods    goodId1    goodMessage1    goodId1
    ...    ${tokenHolder}
    sleep    10
    GetTxByReqId    ${reqId}
    ${result}    GetProof    enterprises    enterprisesId1
    BuiltIn.Should Be Equal As Strings    ${result["Value"]}    enterprisesMessage1
    GetProof    goods    goodId1
    GetProofByCategory    enterprises
    GetProofByCategory    goods
    GetProofByOwner    ${tokenHolder}
    GetProofByReference    goodId1

invoke02
    [Documentation]    ${newAddr}
    ${newAddr}    newAccount
    log    ${newAddr}
    unlockAddr    ${newAddr}    1
    transferP    ${tokenHolder}    ${newAddr}    1000    1    1
    sleep    3
    getBalance    ${newAddr}
    Given Unlock token holder succeed
    ${reqId} =    AddProof    ${tokenHolder}    enterprises    enterprisesId2    enterprisesMessage2    enterprisesId2
    ...    ${newAddr}
    sleep    10
    GetTxByReqId    ${reqId}
    ${reqId} =    AddProof    ${newAddr}    goods    goodId2    goodMessage2    goodId2
    ...    ${newAddr}
    sleep    10
    GetTxByReqId    ${reqId}
    ${result}    GetProof    enterprises    enterprisesId2
    BuiltIn.Should Be Equal As Strings    ${result["Value"]}    enterprisesMessage2
    GetProofByCategory    enterprises
    GetProofByCategory    goods
    GetProofByOwner    ${newAddr}
    GetProofByReference    goodId2
    ${reqId} =    AddProof    ${newAddr}    goods    goodId2    goodMessage22    goodId2
    ...    ${newAddr}
    sleep    10
    GetTxByReqId    ${reqId}
    ${result}    GetProof    goods    goodId2
    BuiltIn.Should Be Equal As Strings    ${result["Value"]}    goodMessage22
    GetProofByCategory    enterprises
    GetProofByCategory    goods
    GetProofByOwner    ${newAddr}
    GetProofByReference    goodId2

invoke03
    [Documentation]    ${newAddr1}
    ${newAddr1}    newAccount
    log    ${newAddr1}
    unlockAddr    ${newAddr1}    1
    transferP    ${tokenHolder}    ${newAddr1}    1000    1    1
    sleep    3
    getBalance    ${newAddr1}
    Given Unlock token holder succeed
    ${reqId} =    AddProof    ${tokenHolder}    enterprises    enterprisesId3    enterprisesMessage3    enterprisesId3
    ...    ${newAddr1}
    sleep    10
    GetTxByReqId    ${reqId}
    ${reqId} =    AddProof    ${newAddr1}    goods    goodId3    goodMessage3    goodId3
    ...    ${newAddr1}
    sleep    10
    GetTxByReqId    ${reqId}
    ${result}    GetProof    enterprises    enterprisesId3
    BuiltIn.Should Be Equal As Strings    ${result["Value"]}    enterprisesMessage3
    GetProofByCategory    enterprises
    GetProofByCategory    goods
    GetProofByOwner    ${newAddr1}
    GetProofByReference    goodId3
    ${reqId} =    DelProof    ${newAddr1}    goods    goodId3
    sleep    10
    GetTxByReqId    ${reqId}
    ${result}    GetProof    goods    goodId3
    BuiltIn.Should Be Equal As Strings    ${result["Value"]}    ${EMPTY}
    ${result}    GetProofByCategory    enterprises
    log    ${result}
    ${result}    GetProofByCategory    goods
    log    ${result[2]}
    GetProofByOwner    ${newAddr1}
    GetProofByReference    goodId3
    #
    ${reqId} =    AddProof    ${newAddr1}    goods    goodId2    goodMessage2    goodId2
    ...    ${newAddr1}
    sleep    10
    GetTxByReqId    ${reqId}
    ${result}    GetProof    goods    goodId2
    BuiltIn.Should Be Equal As Strings    ${result["Value"]}    goodMessage22
    GetProofByReference    goodId2
    GetProofByCategory    goods
    ${reqId} =    DelProof    ${newAddr1}    goods    goodId1
    sleep    10
    GetTxByReqId    ${reqId}
    ${result}    GetProof    goods    goodId1
    BuiltIn.Should Be Equal As Strings    ${result["Value"]}    goodMessage1
    GetProofByReference    goodId1
    GetProofByCategory    goods
    #
    ${admin}    GetAdmin
    BuiltIn.Should Be Equal As Strings    ${admin}    ${tokenHolder}
    ${reqId} =    SetAdmin    ${tokenHolder}    ${newAddr1}
    sleep    10
    GetTxByReqId    ${reqId}
    ${admin}    GetAdmin
    BuiltIn.Should Be Equal As Strings    ${admin}    ${newAddr1}
    ${reqId} =    AddProof    ${newAddr1}    goods    goodId2    ""    goodId2
    ...    ${newAddr1}
    sleep    10
    GetTxByReqId    ${reqId}
    ${result}    GetProof    goods    goodId2
    BuiltIn.Should Be Equal As Strings    ${result["Value"]}    ""
    GetProofByReference    goodId2
    GetProofByCategory    goods
    ${reqId} =    DelProof    ${newAddr1}    goods    goodId1
    sleep    10
    GetTxByReqId    ${reqId}
    ${result}    GetProof    goods    goodId1
    BuiltIn.Should Be Equal As Strings    ${result["Value"]}    ${EMPTY}
    GetProofByReference    goodId1
    GetProofByCategory    goods

stop
    Given Unlock token holder succeed
    ${reqId}=    Then stopContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

*** Keywords ***
InvokeTrace
    [Arguments]    ${addr}    @{args}
    ${respJson}=    invokeContract    ${addr}    ${addr}    0    1    ${gContractId}
    ...    @{args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    request_id
    ${contractId}=    Get From Dictionary    ${result}    contract_id
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

GetTxByReqId
    [Arguments]    ${reqId}
    ${params}    Create List    ${reqId}
    ${respJson}=    sendRpcPost    ${host}    dag_getTxByReqId    ${params}    GetTxByReqId
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    log    ${respJson}
    ${errCode}=    Evaluate    re.findall('\"error_code\":(\\d*)', '${result}')    re
    ${errMsg}=    Evaluate    re.findall('\"error_message\":\"([^"]*)\"', '${result}')    re
    log    ${errCode}=
    log    ${errMsg}=

AddProof
    [Arguments]    ${addr}    ${category}    ${key}    ${value}    ${reference}    ${ownerAddress}
    [Documentation]    调用参数列表：函数名字 addProof，(类别Category，键Key，值Value，引用Reference，所有者OwnerAddress)
    ${args}    Create List    addProof    ${category}    ${key}    ${value}    ${reference}
    ...    ${ownerAddress}
    ${reqId}    InvokeTrace    ${addr}    ${args}
    [Return]    ${reqId}

DelProof
    [Arguments]    ${addr}    ${category}    ${key}
    [Documentation]    调用参数列表：函数名字 delProof，(类别Category，键Key)
    ${args}    Create List    delProof    ${category}    ${key}
    ${reqId}    InvokeTrace    ${addr}    ${args}
    [Return]    ${reqId}

GetProof
    [Arguments]    ${category}    ${key}
    [Documentation]    调用参数列表：函数名字 getProof，(类别Category，键Key)
    ${args}    Create List    getProof    ${category}    ${key}
    ${res}    queryContract    ${gContractId}    ${args}
    log    ${res}
    ${res}    Get From Dictionary    ${res}    result
    ${addressMap}    To Json    ${res}
    [Return]    ${addressMap}

GetProofByCategory
    [Arguments]    ${category}
    [Documentation]    调用参数列表：函数名字 getProofByCategory，(类别Category)
    ${args}    Create List    getProofByCategory    ${category}
    ${res}    queryContract    ${gContractId}    ${args}
    log    ${res}
    ${res}    Get From Dictionary    ${res}    result
    ${addressMap}    To Json    ${res}
    [Return]    ${addressMap}

GetProofByOwner
    [Arguments]    ${ownerAddress}
    [Documentation]    调用参数列表：函数名字 getProofByOwner
    ${args}    Create List    getProofByOwner    ${ownerAddress}
    ${res}    queryContract    ${gContractId}    ${args}
    log    ${res}

GetProofByReference
    [Arguments]    ${reference}
    [Documentation]    调用参数列表：函数名字 getProofByReference
    ${args}    Create List    getProofByReference    ${reference}
    ${res}    queryContract    ${gContractId}    ${args}
    log    ${res}

SetAdmin
    [Arguments]    ${oldAdmin}    ${newAdmin}
    [Documentation]    调用参数列表：函数名字 setAdmin，PTN地址 \ 例如，P19z4r7G9MpZtaYMZcATWinTwXeGBj7fWTd
    ${args}    Create List    setAdmin    ${newAdmin}
    ${reqId}    InvokeTrace    ${oldAdmin}    ${args}
    [Return]    ${reqId}

newAccount
    ${param}    Create List    1
    ${result}    post    personal_newAccount    personal_newAccount    ${param}
    log    ${result}
    [Return]    ${result}

post
    [Arguments]    ${method}    ${alias}    ${params}
    ${header}    Create Dictionary    Content-Type=application/json
    ${data}    Create Dictionary    jsonrpc=2.0    method=${method}    params=${params}    id=1
    Create Session    ${alias}    ${host}    #    http://127.0.0.1:8645    http://192.168.44.128:8645
    ${resp}    Post Request    ${alias}    ${host}    data=${data}    headers=${header}
    ${respJson}    To Json    ${resp.content}
    Dictionary Should Contain Key    ${respJson}    result
    ${res}    Get From Dictionary    ${respJson}    result
    [Return]    ${res}

GetAdmin
    ${args}    Create List    getAdmin
    ${res}    queryContract    ${gContractId}    ${args}
    log    ${res}
    ${addressMap}    Get From Dictionary    ${res}    result
    [Return]    ${addressMap}

unlockAddr
    [Arguments]    ${addr}    ${pwd}
    ${param}    Create List    ${addr}    ${pwd}
    ${result}    post    personal_unlockAccount    personal_unlockAccount    ${param}
    log    ${result}
    Should Be True    ${result}

transferP
    [Arguments]    ${fromAddr}    ${toAddr}    ${amount}    ${fee}    ${pwd}
    ${param}    Create List    ${fromAddr}    ${toAddr}    ${amount}    ${fee}    ${null}
    ...    ${pwd}
    ${result}    post    wallet_transferPtn    wallet_transferPtn    ${param}
    log    ${result}

getBalance
    [Arguments]    ${addr}
    ${param}    Create List    ${addr}
    ${result}    post    wallet_getBalance    wallet_getBalance    ${param}
    log    ${result}
    ${amount}    Set Variable    ${result["PTN"]}
    [Return]    ${amount}
