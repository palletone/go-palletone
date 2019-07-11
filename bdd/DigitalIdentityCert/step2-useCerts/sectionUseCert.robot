*** Settings ***
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot

*** Test Cases ***
sectionUseCert
    Given section unlock account succed
    When section uses debug contract to test getRequesterCert without error
    And Wait for transaction being packaged
    And section uses debug contract to test checkRequesterCert without error
    And Wait for transaction being packaged
    Then print out: section has authority to use this cert

*** Keywords ***
section unlock account succed
    Log    "section unlock account succed"
    ${respJson}=    unlockAccount    ${sectionCertHolder}
    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${respJson["result"]}    ${true}

section uses debug contract to test getRequesterCert without error
    Log    "section uses debug contract to test getRequesterCert without error"
    ${args}=    Create List    ${getRequesterCertMethod}
    ${params}=    genInvoketxParams    ${sectionCertHolder}    ${sectionCertHolder}    1    1    ${debugContractAddr}
    ...    ${args}    ${sectionCertID}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    getRequesterCert
    Dictionary Should Contain Key    ${respJson}    result

section uses debug contract to test checkRequesterCert without error
    Log    "section uses debug contract to test checkRequesterCert without error"
    ${args}=    Create List    ${checkRequesterCertMethod}
    ${params}=    genInvoketxParams    ${sectionCertHolder}    ${sectionCertHolder}    1    1    ${debugContractAddr}
    ...    ${args}    ${sectionCertID}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    checkRequesterCert
    Dictionary Should Contain Key    ${respJson}    result

print out: section has authority to use this cert
    Log    "print out: section has authority to use this cert"
