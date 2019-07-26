*** Settings ***
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot
Library           DateTime
Library           String

*** Test Cases ***
2ndCAInvokePowerCert
    Given ca unlock his account succed
    When ca revoke power certificate succed
    And Wait for transaction being packaged
    And ca can query his issued crl file
    Then power and section certificate revocation time is before now

*** Keywords ***
ca unlock his account succed
    Log    "ca unlock his account succed"
    ${respJson}=    unlockAccount    ${caCertHolder}
    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${respJson["result"]}    ${true}

ca revoke power certificate succed
    Log    "ca revoke power certificate succed"
    ${args}=    Create List    ${addCRLMethod}    ${caRevokePowerCRLBytes}
    ${params}=    genInvoketxParams    ${caCertHolder}    ${caCertHolder}    1    1    ${certContractAddr}
    ...    ${args}    ${null}
    ${respJson}=    sendRpcPost    ${invokeMethod}    ${params}    addCRL
    Dictionary Should Contain Key    ${respJson}    result

ca can query his issued crl file
    Log    "ca can query his issued crl file"
    ${args}=    Create List    ${queryCRLMethod}    ${caCertHolder}
    ${params}=    Create List    ${certContractAddr}    ${args}    ${0}
    ${respJson}=    sendRpcPost    ${queryMethod}    ${params}    queryCRL
    Dictionary Should Contain Key    ${respJson}    result
    ${bytes}=    Evaluate    ${respJson['result']}
    Length Should Be    ${bytes}    1

power and section certificate revocation time is before now
    power certificate revocation time is before now
    section certificate revocation time is before now

power certificate revocation time is before now
    ${args}=    Create List    ${getHolderCertMethod}    ${powerCertHolder}
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
    ...    ELSE    Fail    ca invoke power certificate failed

section certificate revocation time is before now
    ${args}=    Create List    ${getHolderCertMethod}    ${sectionCertHolder}
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
    ...    ELSE    Fail    ca invoke power certificate failed
