*** Settings ***
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot

*** Test Cases ***
caUseCert
    Given ca unlock account succed
    When ca uses debug contract to test getRequesterCert without error
    And Wait for transaction being packaged
    And ca uses debug contract to test checkRequesterCert without error
    And Wait for transaction being packaged
    Then print out: ca has authority to use this cert

*** Keywords ***
ca unlock account succed
    Log    "ca unlock account succed"
    ${respJson}=    unlockAccount    ${caCertHolder}
    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${respJson["result"]}    ${true}

ca uses debug contract to test getRequesterCert without error
    Log    "ca uses debug contract to test getRequesterCert without error"
    ${args}=    Create List    ${getRequesterCertMethod}
    ${params}=    genInvoketxParams    ${caCertHolder}    ${caCertHolder}    1    1    ${debugContractAddr}
    ...    ${args}    ${caCertID}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    getRequesterCert
    Dictionary Should Contain Key    ${respJson}    result

ca uses debug contract to test checkRequesterCert without error
    Log    "ca uses debug contract to test checkRequesterCert without error"
    ${args}=    Create List    ${checkRequesterCertMethod}
    ${params}=    genInvoketxParams    ${caCertHolder}    ${caCertHolder}    1    1    ${debugContractAddr}
    ...    ${args}    ${caCertID}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    checkRequesterCert
    Dictionary Should Contain Key    ${respJson}    result

print out: ca has authority to use this cert
    Log    "print out: ca has authority to use this cert"
