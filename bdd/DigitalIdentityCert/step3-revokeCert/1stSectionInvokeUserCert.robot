*** Settings ***
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot
Library           DateTime
Library           String

*** Test Cases ***
1stSectionInvokeUserCert
    Given section unlock his account succed
    When section revoke user certificate succed
    And Wait for transaction being packaged
    And section can query his issued crl file
    Then user certificate revocation time is before now

*** Keywords ***
section unlock his account succed
    Log    "section unlock his account succed"
    ${respJson}=    unlockAccount    ${sectionCertHolder}
    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${respJson["result"]}    ${true}

section revoke user certificate succed
    Log    "section revoke user certificate succed"
    ${args}=    Create List    ${addCRLMethod}    ${sectionRevokeUserCRLBytes}
    ${params}=    genInvoketxParams    ${sectionCertHolder}    ${sectionCertHolder}    1    1    ${certContractAddr}
    ...    ${args}    ${null}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    addCRL
    Dictionary Should Contain Key    ${respJson}    result

section can query his issued crl file
    Log    "section can query his issued crl file"
    ${args}=    Create List    ${queryCRLMethod}    ${sectionCertHolder}
    ${params}=    Create List    ${certContractAddr}    ${args}    ${0}
    ${respJson}=    sendRpcPost    ${queryMethod}    ${params}    queryCRL
    Dictionary Should Contain Key    ${respJson}    result
    ${bytes}=    Evaluate    ${respJson['result']}
    Length Should Be    ${bytes}    1

user certificate revocation time is before now
    ${args}=    Create List    ${getHolderCertMethod}    ${userCertHolder}
    ${params}=    Create List    ${certContractAddr}    ${args}    ${0}
    ${respJson}=    sendRpcPost    ${queryMethod}    ${params}    queryCert
    Dictionary Should Contain Key    ${respJson}    result
    ${resultDict}=    Evaluate    ${respJson["result"]}
    Dictionary Should Contain Key    ${resultDict}    IntermediateCertIDs
    Length Should Be    ${resultDict['IntermediateCertIDs']}    1
    Dictionary Should Contain Key    ${resultDict['IntermediateCertIDs'][0]}    CertID
    ${now}=    Get Current Date    UTC
    ${words}=    Split String    ${resultDict['IntermediateCertIDs'][0]['RecovationTime']}    ${SPACE}
    Length Should Be    ${words}    4
    ${sRevocationTime}=    catenate    ${words[0]}    ${words[1]}
    ${sRevocationTime}=    catenate    SEPARATOR=    ${sRevocationTime}    .000
    ${revocationTime}=    Convert Date    ${sRevocationTime}
    Run Keyword If    '${now}'>'${revocationTime}'    log    1
    ...    ELSE    Fail    section invoke user certificate failed
