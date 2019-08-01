*** Settings ***
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot
Resource          ../setups.robot

*** Test Cases ***
powerIssueSection
    Given power unlock its account succeed
    When power issues intermediate certificate name cert2 to section succeed
    and Wait for transaction being packaged
    Then section can query his certificate in db

*** Keywords ***
power unlock its account succeed
    Log    "power unlock its account succeed"
    ${respJson}=    unlockAccount    ${powerCertHolder}
    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${respJson["result"]}    ${true}

power issues intermediate certificate name cert2 to section succeed
    Log    "power issues intermediate certificate name cert2 to section succeed"
    ${args}=    Create List    addServerCert    ${sectionCertHolder}    ${sectionCertBytes}
    ${params}=    genInvoketxParams    ${powerCertHolder}    ${powerCertHolder}    1    1    ${certContractAddr}
    ...    ${args}    ${null}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    addServerCert
    Dictionary Should Contain Key    ${respJson}    result

section can query his certificate in db
    Log    "section can query his certificate in ledger"
    ${args}=    Create List    ${getHolderCertMethod}    ${sectionCertHolder}
    ${params}=    Create List    ${certContractAddr}    ${args}    ${0}
    ${respJson}=    sendRpcPost    ${queryMethod}    ${params}    queryCert
    Dictionary Should Contain Key    ${respJson}    result
    ${resultDict}=    Evaluate    ${respJson["result"]}
    Dictionary Should Contain Key    ${resultDict}    IntermediateCertIDs
    Length Should Be    ${resultDict['IntermediateCertIDs']}    1
    Dictionary Should Contain Key    ${resultDict['IntermediateCertIDs'][0]}    CertID
    Should Be Equal    ${resultDict['IntermediateCertIDs'][0]['CertID']}    ${sectionCertID}
