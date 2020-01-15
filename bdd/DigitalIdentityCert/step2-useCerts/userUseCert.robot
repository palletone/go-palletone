*** Settings ***
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot

*** Test Cases ***
userUseCert
    Given user unlock account succed
    When user uses debug contract to test getRequesterCert without error
    And Wait for transaction being packaged
    And user uses debug contract to test checkRequesterCert without error
    And Wait for transaction being packaged
    Then print out: user has authority to use this cert

*** Keywords ***
user unlock account succed
    Log    "user unlock account succed"
    ${respJson}=    unlockAccount    ${userCertHolder}
    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${respJson["result"]}    ${true}

user uses debug contract to test getRequesterCert without error
    Log    "user uses debug contract to test getRequesterCert without error"
    ${args}=    Create List    ${getRequesterCertMethod}
    ${params}=    genInvoketxParams    ${userCertHolder}    ${userCertHolder}    1    1    ${debugContractAddr}
    ...    ${args}    ${userCertID}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    getRequesterCert
    Dictionary Should Contain Key    ${respJson}    result

user uses debug contract to test checkRequesterCert without error
    Log    "user uses debug contract to test checkRequesterCert without error"
    ${args}=    Create List    ${checkRequesterCertMethod}
    ${params}=    genInvoketxParams    ${userCertHolder}    ${userCertHolder}    1    1    ${debugContractAddr}
    ...    ${args}    ${userCertID}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    checkRequesterCert
    Dictionary Should Contain Key    ${respJson}    result

print out: user has authority to use this cert
    Log    "print out: user has authority to use this cert"
