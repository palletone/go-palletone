*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot
Library           BuiltIn

*** Test Cases ***
InstallTestshimucTpl
    Given Unlock token holder succeed
    ${reqId} =    When User installs contract template    github.com/palletone/go-palletone/contracts/example/go/testshimuc    testshimuc
    And wait for transaction being packaged
    Then Wait for unit abount contract to be confirmed by unit height    ${reqId}

DeployTestshimuc
    Given Unlock token holder succeed
    ${reqId} =    When User deploys contract
    And wait for transaction being packaged
    Then Wait for unit abount contract to be confirmed by unit height    ${reqId}

testGetArgs
    Given Unlock token holder succeed
    ${args} =    When Generate arguments for test
    ${res}=    And User get arguments    testGetArgs    ${args}
    Then Check arguments    ${res}

*** Keywords ***
Generate arguments for test
    ${args}=    Create List    aa    bb    cc
    [Return]    ${args}

Check arguments
    [Arguments]    ${args}
    ${origin}=    Create List    aa bb cc
    Lists Should Be Equal    ${origin}    ${args}

User get arguments
    [Arguments]    ${getmethod}    ${args}
    ${args}=    Create List    ${getmethod}    ${args}
    ${respJson}=    queryContract    ${gContractId}    ${args}
    Dictionary Should Contain Key    ${respJson}    result
    ${result}=    Get From Dictionary    ${respJson}    result
