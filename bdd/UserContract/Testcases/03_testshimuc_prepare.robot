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