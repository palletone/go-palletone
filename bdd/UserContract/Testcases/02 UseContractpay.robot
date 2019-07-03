*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot
Library           BuiltIn

*** Test Cases ***
InstallContractpayTpl
    Given Unlock token holder succeed
    ${reqId} =    When User installs contract template
    And wait for transaction being packaged
    Then Wait for unit abount contract to be confirmed by unit height    ${reqId}

DeployContract
    Given Unlock token holder succeed
    ${reqId} =    When User deploys contract
    And wait for transaction being packaged
    Then Wait for unit abount contract to be confirmed by unit height    ${reqId}

PutStatus
    Given Unlock token holder succeed
    ${reqId} =    When User put status into contractpay    put
    And wait for transaction being packaged
    And Wait for unit abount contract to be confirmed by unit height    ${reqId}
    Then Get status from contractpay    get    a    aa

Paystate1
    Given Unlock token holder succeed
    ${reqId} =    When User put status into contractpay    paystate1
    And wait for transaction being packaged
    And Wait for unit abount contract to be confirmed by unit height    ${reqId}
    Then Get status from contractpay    get    paystate1    paystate1

Paystate2
    Given Unlock token holder succeed
    ${reqId} =    When User put status into contractpay    paystate2
    And wait for transaction being packaged
    And Wait for unit abount contract to be confirmed by unit height    ${reqId}
    Then Get status from contractpay    get    paystate2    paystate2

Payout
    Given Unlock token holder succeed
    When User transfer PTN to contractpay
    And wait for transaction being packaged
    And Query contract balance
    ${newAddr}    ${reqId}=    And Use contractpay to transfer PTN to user2
    And wait for transaction being packaged
    And Wait for unit abount contract to be confirmed by unit height    ${reqId}
    Then Query user2 balance    ${newAddr}

*** Keywords ***
Unlock token holder succeed
    unlockAccount    ${tokenHolder}

User installs contract template
    ${respJson}=    installContractTpl    ${tokenHolder}    ${tokenHolder}    100    100    jury06
    ...    github.com/palletone/go-palletone/contracts/example/go/contractpay    example
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${tplId}=    Get From Dictionary    ${result}    tplId
    Run Keyword If    '${tplId}'=='${EMPTY}'    Fail    "Install Contract Error"
    Set Global Variable    ${gTplId}    ${tplId}
    [Return]    ${reqId}

User deploys contract
    ${args}=    Create List    deploy
    ${respJson}=    deployContract    ${tokenHolder}    ${tokenHolder}    1000    10    ${gTplId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Run Keyword If    '${contractId}'=='${EMPTY}'    Fail    "Deploy Contract Error"
    Set Global Variable    ${gContractId}    ${contractId}
    [Return]    ${reqId}

User put status into contractpay
    [Arguments]    ${putmethod}
    ${args}=    Create List    ${putmethod}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

Get status from contractpay
    [Arguments]    ${getmethod}    ${name}    ${result}
    ${args}=    Create List    ${getmethod}    ${name}
    ${respJson}=    queryContract    ${gContractId}    ${args}
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Should Be Equal    ${result}    ${result}

User transfer PTN to contractpay
    transferPtnTo    ${gContractId}    10000
    wait for transaction being packaged

Query contract balance
    Set Global Variable    ${gasToken}    PTN
    ${amount}=    getBalance    ${gContractId}
    Should Be Equal    ${amount}    10000
    Log    ${amount}

Use contractpay to transfer PTN to user2
    # create account
    ${newAddr}=    newAccount
    Log    ${newAddr}
    ${args}=    Create List    payout    ${newAddr}    PTN    100
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
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
    Should Be Equal    ${ramount}    100
    Should Be Equal    ${rasset}    PTN
    Should Be Equal    ${raddr}    ${addr}
