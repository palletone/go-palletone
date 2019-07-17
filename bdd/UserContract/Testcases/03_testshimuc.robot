*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot
Library           BuiltIn
Library           Collections

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
    ${reqId}=    When User put state    testPutState    state2    state2
    And Wait for transaction being packaged
    And Wait for unit about contract to be confirmed by unit height    ${reqId}
    # -------- put global state should be error ----------
    ${reqId}=    And User put state    testPutGlobalState    gState1    gState1
    And Wait for transaction being packaged
    ${errCode}    ${errMsg}=    And Wait for unit about contract to be confirmed by unit height    ${reqId}
    Should Be Equal    ${errMsg}    Chaincode Error:Only system contract can call this function.
    # -------- query contract state --------------
    Then User query state    testGetState    state1    state1    str    ${null}
    Then User query state    testGetState    state2    state2    str    ${null}
    And User query state    testGetContractState    state1    state1    str    ${gContractId}
    And User query state    testGetContractState    state2    state2    str    ${gContractId}
    And User query state    testGetContractState    gState1    ${EMPTY}    str    ${gContractId}
    And User query state    testGetGlobalState    gState1    ${EMPTY}    str    ${null}
    ${allState}=    And Create Dictionary    state1    state1    state2    state2
    And User query state    testGetStateByPrefix    state    ${allState}    dict    ${null}
    ${allState}=    And Create Dictionary    paystate0    paystate0    state1    state1    state2
    ...    state2
    And User query state    testGetContractAllState    ${null}    ${allState}    dict    ${null}

DelState
    Given Unlock token holder succeed
    ${reqId}=    When User delete state    testDelState    state1
    And Wait for transaction being packaged
    And Wait for unit about contract to be confirmed by unit height    ${reqId}
    ${reqId}=    And User delete state    testDelState    state2
    And Wait for transaction being packaged
    And Wait for unit about contract to be confirmed by unit height    ${reqId}
    Then User query state    testGetState    state1    ${EMPTY}    str    ${null}
    And User query state    testGetGlobalState    state2    ${EMPTY}    str    ${null}
    And User query state    testGetContractState    state1    ${EMPTY}    str    ${gContractId}
    And User query state    testGetContractState    state2    ${EMPTY}    str    ${gContractId}
    And User query state    testGetStateByPrefix    state    ${null}    dict    ${null}
    ${allState}=    And Create Dictionary    paystate0    paystate0
    And User query state    testGetContractAllState    ${null}    ${allState}    dict    ${null}

HandleToken
    Given Unlock token holder succeed
    ${reqId}=    When User define token
    ${reqId}=    And User supply token
    ${reqId}=    And User pay out token
    Then User query balance

Stop testshimuc contract
    Given Unlock token holder succeed
    ${reqId}=    Then stopContract    ${tokenHolder}    ${tokenHolder}    100    1    ${gContractId}
    And Wait for transaction being packaged
    And Wait for unit about contract to be confirmed by unit height    ${reqId}

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
    [Arguments]    ${getmethod}    ${name}    ${exceptedResult}    ${resType}    ${contractId}
    ${args}=    Run Keyword If    '${contractId}'=='${null}'    Create List    ${getmethod}    ${name}
    ...    ELSE    Create List    ${getmethod}    ${contractId}    ${name}
    ${respJson}=    queryContract    ${gContractId}    ${args}
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
    ${resDict}=    Run Keyword If    '${resType}'=='dict'    To Json    ${result}
    Run Keyword If    '${resType}'=='str'    Should Be Equal    '${result}'    '${exceptedResult}'
    ...    ELSE    Dictionaries Should Be Equal    ${resDict}    ${exceptedResult}
