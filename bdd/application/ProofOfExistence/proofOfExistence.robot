*** Settings ***
Test Setup        beforeVote
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot
Resource          ../setups.robot

*** Test Cases ***
mediatorvote
    Given user unlock its account succeed
    When create proof existence
    and wait transaction packaged
    Then check proof existence

*** Keywords ***
user unlock its account succeed
    ${respJson}=    unlockAccount    ${userAccount}
    log    ${respJson}
    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${respJson["result"]}    ${true}

create proof existence
    Log    create proof existence
    ${args}=    Create List
    ${params}=    Create List    ${userAccount}    ${maindata}    ${extradata}    ${reference}    ${fee}
    ${resp}=    sendRpcPost    ${createProofExistence}    ${params}    create proof existence
    ${res}    set variable    ${resp["result"]}
    log    ${res}    INFO

wait transaction packaged
    Log    wait for transaction being packaged
    Sleep    15

check proof existence
    Log    check proof existence
    ${args}=    Create List
    ${params}=    Create List    ${reference}
    ${resp}=    sendRpcPost    ${checkProofExistence}    ${params}    check proof existence
    ${res}=    Get From Dictionary    ${resp}    result
    ${refs}=    Get From List    ${res}    0
    ${ref}    set variable    ${refs["reference"]}
    Should Be Equal    ${ref}    ${reference}
    log    ${ref}    INFO
