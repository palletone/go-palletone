*** Settings ***
Resource          pubVariables.robot
Resource          pubFuncs.robot
Library           Collections

*** Keywords ***
beforeIssueCerts
    queryCAHolder
    newAccounts
    transferPTNToIntermediateCertUsers

newAccounts
    # create power intermediate certificate holder
    ${user1}=    newAccount
    Set Global Variable    ${powerCertHolder}    ${user1}
    # create section intermediate certificate holder
    ${user2}=    newAccount
    Set Global Variable    ${sectionCertHolder}    ${user2}
    # create user certificate holder
    ${user3}=    newAccount
    Set Global Variable    ${userCertHolder}    ${user3}

queryCAHolder
    ${args}=    Create List    getRootCAHoler
    ${params}=    Create List    ${certContractAddr}    ${args}   ${0}
    ${respJson}=    sendRpcPost    ${queryMethod}    ${params}    getCAHolder
    Dictionary Should Contain Key    ${respJson}    result
    Set Global Variable    ${caCertHolder}    ${respJson["result"]}
    Set Global Variable    ${tokenHolder}    ${caCertHolder}

transferPTNToIntermediateCertUsers
    # transfer PTN to power intermediate certificate holder
    transferPTN    ${powerCertHolder}
    Log    wait for tx being packaged into unit
    sleep    6    # should sleep, because transaction has not been packaged into unit
    ${balance}=    getBalance    ${powerCertHolder}
    Should Be Equal    ${balance}    ${amount}
    # transfer PTN to section intermediate certificate holder
    transferPTN    ${sectionCertHolder}
    Log    wait for tx being packaged into unit
    sleep    6    # should sleep, because transaction has not been packaged into unit
    ${balance}=    getBalance    ${sectionCertHolder}
    Should Be Equal    ${balance}    ${amount}
    # transfer PTN to user certificate holder
    transferPTN    ${userCertHolder}
    Log    wait for tx being packaged into unit
    sleep    6    # should sleep, because transaction has not been packaged into unit
    ${balance}=    getBalance    ${userCertHolder}
    Should Be Equal    ${balance}    ${amount}
