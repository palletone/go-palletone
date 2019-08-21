*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot
Library           BuiltIn
Library           Collections

*** Test Cases ***
InstallContractpayTpl
    Given Unlock token holder succeed
    ${reqId} =    When User installs contract template    github.com/palletone/go-palletone/contracts/example/go/contractpay    example
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

DeployContract
    Given Unlock token holder succeed
    ${reqId} =    When User deploys contract
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

PutStatus
    Given Unlock token holder succeed
    ${reqId} =    When User put status into contractpay    put
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    #Then Get status from contractpay    get    a    aa

Paystate1
    Given Unlock token holder succeed
    ${reqId} =    When User put status into contractpay    paystate1
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    #Then Get status from contractpay    get    paystate1    paystate1

Paystate2
    Given Unlock token holder succeed
    ${reqId} =    When User put status into contractpay    paystate2
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    #Then Get status from contractpay    get    paystate2    paystate2

Payout
    Given Unlock token holder succeed
    When User transfer PTN to contractpay
    And Wait for transaction being packaged
    And Query contract balance
    ${newAddr}    ${reqId}=    And Use contractpay to transfer PTN to user2
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}
    Then Query user2 balance    ${newAddr}

Stop contractpay contract
    Given Unlock token holder succeed
    ${reqId}=   Then stopContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    And Wait for unit about contract to be confirmed by unit height    ${reqId}    ${true}

*** Keywords ***
User put status into contractpay
    [Arguments]    ${putmethod}
    ${args}=    Create List    ${putmethod}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

Get status from contractpay
    [Arguments]    ${getmethod}    ${name}    ${exceptedResult}
    ${args}=    Create List    ${getmethod}    ${name}
    ${respJson}=    queryContract    ${gContractId}    ${args}
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Should Be Equal    ${result}    ${exceptedResult}

User transfer PTN to contractpay
    transferPtnTo    ${gContractId}    10000
    Wait for transaction being packaged

Query contract balance
    ${amount}=    getBalance    ${gContractId}  PTN
    Should Be Equal    ${amount}    10000
    Log    ${amount}

Use contractpay to transfer PTN to user2
    # create account
    ${newAddr}=    newAccount
    Log    ${newAddr}
    ${args}=    Create List    payout    ${newAddr}    PTN    100
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    100    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    [Return]    ${newAddr}    ${reqId}

Query user2 balance
    [Arguments]    ${addr}
    ${args}=    Create List    balance    ${addr}
    ${respJson}=    queryContract    ${gContractId}    ${args}
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    ${result}=    To Json    ${result}
    ${balance}=    Get From List    ${result}    0
    ${ramount}=    Get From Dictionary    ${balance}    amount
    ${rasset}=    Get From Dictionary    ${balance}    asset
    ${raddr}=    Get From Dictionary    ${balance}    address
    Should Be Equal    ${ramount}    ${100}
    Should Be Equal    ${rasset}    PTN
    Should Be Equal    ${raddr}    ${addr}
