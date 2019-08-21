*** Settings ***
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot

*** Test Cases ***
powerUseCert
    Given power unlock account succed
    When power uses debug contract to test getRequesterCert without error
    And Wait for transaction being packaged
    And power uses debug contract to test checkRequesterCert without error
    And Wait for transaction being packaged
    Then print out: power has authority to use this cert

*** Keywords ***
power unlock account succed
    Log    "power unlock account succed"
    ${respJson}=    unlockAccount    ${powerCertHolder}
    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${respJson["result"]}    ${true}

power uses debug contract to test getRequesterCert without error
    Log    "power uses debug contract to test getRequesterCert without error"
    ${args}=    Create List    ${getRequesterCertMethod}
    ${params}=    genInvoketxParams    ${powerCertHolder}    ${powerCertHolder}    1    1    ${debugContractAddr}
    ...    ${args}    ${powerCertID}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    getRequesterCert
    Dictionary Should Contain Key    ${respJson}    result

power uses debug contract to test checkRequesterCert without error
    Log    "power uses debug contract to test checkRequesterCert without error"
    ${args}=    Create List    ${checkRequesterCertMethod}
    ${params}=    genInvoketxParams    ${powerCertHolder}    ${powerCertHolder}    1    1    ${debugContractAddr}
    ...    ${args}    ${powerCertID}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    checkRequesterCert
    Dictionary Should Contain Key    ${respJson}    result

print out: power has authority to use this cert
    Log    "print out: power has authority to use this cert"
