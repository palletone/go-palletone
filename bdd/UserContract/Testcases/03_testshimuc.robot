*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot
Library           BuiltIn

*** Test Cases ***
InstallTestshimucTpl
    Given Unlock token holder succeed
    ${reqId} =    When User installs contract template    github.com/palletone/go-palletone/contracts/example/go/testshimuc    testshimuc
    And Wait for transaction being packaged
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}

DeployTestshimuc
    Given Unlock token holder succeed
    ${reqId} =    When User deploys contract
    And Wait for transaction being packaged
    Then Wait for unit about contract to be confirmed by unit height    ${reqId}

AddState
    Given Unlock token holder succeed
    ${reqId}=    When User put state    testPutState    state1    state1
    And Wait for transaction being packaged
    And Wait for unit about contract to be confirmed by unit height    ${reqId}
    ${reqId}=    And User put state    testPutGlobalState    state2    state2
    And Wait for transaction being packaged
    And Wait for unit about contract to be confirmed by unit height    ${reqId}
    Then User query state    testGetState    state1    state1    str
    And User query state    testGetGlobalState    state2    state2    str
    And User query state    testGetContractState    state1    state1    str
    And User query state    testGetContractState    state2    state2    str
    ${allState}=    And Create Dictionary    state1    state1    state2    state2
    And User query state    testGetStateByPrefix    state    ${allState}    dic
    And User query state    testGetContractAllState    ${null}    ${allState}    dic

DelState
    Given Unlock token holder succeed
    ${reqId}=    When User delete state    testDelState    state1
    And Wait for transaction being packaged    ${reqId}
    ${reqId}=    And User delete state    testDelGlobalState    state2
    And Wait for transaction being packaged    ${reqId}
    Then User query state    testGetState    state1    ${null}
    And User query state    testGetGlobalState    state2    ${null}
    And User query state    testGetContractState    state1    ${null}
    And User query state    testGetContractState    state2    ${null}
    ${allState}=    And Create Dictionary    state1    state1    state2    state2
    And User query state    testGetStateByPrefix    state    ${null}    dic
    And User query state    testGetContractAllState    ${null}    ${null}    dic

HandleToken
    Given Unlock token holder succeed
    ${reqId}=    When User define token
    ${reqId}=    And User supply token
    ${reqId}=    And User pay out token
    Then User query balance

*** Keywords ***
User put state
    [Arguments]    ${method}    ${key}    ${value}
    ${args}=    Create List    ${method}    ${key}    ${value}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

User delete state
    [Arguments]    ${method}    ${key}
    ${args}=    Create List    ${method}    ${key}
    ${respJson}=    invokeContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
    ...    ${args}
    ${result}=    Get From Dictionary    ${respJson}    result
    ${reqId}=    Get From Dictionary    ${result}    reqId
    ${contractId}=    Get From Dictionary    ${result}    ContractId
    Should Be Equal    ${gContractId}    ${contractId}
    [Return]    ${reqId}

User query state
    [Arguments]    ${getmethod}    ${name}    ${exceptedResult}    ${resType}
    ${args}=    Create List    ${getmethod}    ${name}
    ${respJson}=    queryContract    ${gContractId}    ${args}
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    Run Keyword If    '${resType}'=='str'    Should Be Equal    ${result}    ${exceptedResult}
    ...    ELSE    Dictionaries Should Be Equal    ${result}    ${exceptedResult}
