*** Settings ***
Test Setup        beforeVote
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot
Resource          ../setups.robot

*** Test Cases ***
mediatorvote
    Given user unlock its account succeed
    When check mediator list
    and wait for transaction being packaged
    and check mediator list
    and vote mediator account
    Then view vote results

*** Keywords ***
user unlock its account succeed
    Log    " user unlock its account succeed"
    ${respJson}=    unlockAccount    ${userAccount}
    log    ${respJson}
    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${respJson["result"]}    ${true}

check mediator list
    Log    "check mediator list succeed"
    ${args}=    Create List
    ${params}=    Create List
    ${resp}=    sendRpcPost    ${checkMediatorList}    ${params}    check mediator list
    ${mediatorAddrs}=    Get From Dictionary    ${resp}    result
    ${firstAddr}=    Get From List    ${mediatorAddrs}    0
    ${secondAddr}=    Get From List    ${mediatorAddrs}    1
    Set Global Variable    ${mediatorHolder1}    ${firstAddr}
    Set Global Variable    ${mediatorHolder2}    ${secondAddr}

vote mediator account
    #Log    "vote mediator account    succeed"
    ${args}=    Create List    ${mediatorHolder1}
    ${params}=    Create List    ${userAccount}    ${args}
    ${resp}=    sendRpcPost    ${voteMediator}    ${params}    vote mediator
    log    ${resp}

view vote results
    Log    wait for transaction being packaged
    Sleep    5
    ${args}=    Create List
    ${params}=    Create List
    ${resp}=    sendRpcPost    ${mediatorVoteResults}    ${params}    view mediator results
    ${resultAddrs}=    Get From Dictionary    ${resp}    result
    Dictionary Should Contain Key    ${resp}    result
    ${voteResult}    set variable    ${resultAddrs['${mediatorHolder1}']}
    Set Global Variable    ${mediator1Result}    ${voteResult}
    run keyword if    ${voteResult}!=0    log    success
    Log    ${mediator1Result}    INFO
