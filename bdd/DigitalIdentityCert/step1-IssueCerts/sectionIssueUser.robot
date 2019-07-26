*** Settings ***
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot
Resource          ../setups.robot

*** Test Cases ***
sectionIssueUser
    Given section unlock its account succeed
    When section issues intermediate certificate name cert2 to user succeed
    and Wait for transaction being packaged
    Then user can query his certificate in db

*** Keywords ***
section unlock its account succeed
    Log    "section unlock its account succeed"
    ${respJson}=    unlockAccount    ${sectionCertHolder}
    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${respJson["result"]}    ${true}

section issues intermediate certificate name cert2 to user succeed
    Log    "section issues intermediate certificate name cert2 to user succeed"
    ${args}=    Create List    addServerCert    ${userCertHolder}    ${userCertBytes}
    ${params}=    genInvoketxParams    ${sectionCertHolder}    ${sectionCertHolder}    1    1    ${certContractAddr}
    ...    ${args}    ${null}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    addServerCert
    Dictionary Should Contain Key    ${respJson}    result

user can query his certificate in db
    Log    "user can query his certificate in db"
    ${args}=    Create List    ${getHolderCertMethod}    ${userCertHolder}
    ${params}=    Create List    ${certContractAddr}    ${args}    ${0}
    ${respJson}=    sendRpcPost    ${queryMethod}    ${params}    queryCert
    Dictionary Should Contain Key    ${respJson}    result
    ${resultDict}=    Evaluate    ${respJson["result"]}
    Dictionary Should Contain Key    ${resultDict}    IntermediateCertIDs
    Length Should Be    ${resultDict['IntermediateCertIDs']}    1
    Dictionary Should Contain Key    ${resultDict['IntermediateCertIDs'][0]}    CertID
    Should Be Equal    ${resultDict['IntermediateCertIDs'][0]['CertID']}    ${userCertID}
