*** Settings ***
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot
Resource          ../setups.robot

*** Test Cases ***
mediatorvote
    Given view vote results
    When wait for the change of mediator
    and choose the winner of mediator
    Then determine if the change mediator

*** Keywords ***
view vote results
    Log    wait for transaction being packaged
    Sleep    5
    ${args}=    Create List
    ${params}=    Create List
    ${resp}=    sendRpcPost    ${mediatorVoteResults}    ${params}    view mediator results
    ${resultAddrs}=    Get From Dictionary    ${resp}    result
    Dictionary Should Contain Key    ${resp}    result
    ${voteResult1}    set variable    ${resultAddrs['${mediatorHolder1}']}
    ${voteResult2}    set variable    ${resultAddrs['${mediatorHolder2}']}
    ${voteResult3}    set variable    ${resultAddrs['${mediatorHolder3}']}
    ${voteResult4}    set variable    ${resultAddrs['${mediatorHolder4}']}
    ${voteResult5}    set variable    ${resultAddrs['${mediatorHolder5}']}
    Set Global Variable    ${mediator1Result}    ${voteResult1}
    Set Global Variable    ${mediator2Result}    ${voteResult2}
    Set Global Variable    ${mediator3Result}    ${voteResult3}
    Set Global Variable    ${mediator4Result}    ${voteResult4}
    Set Global Variable    ${mediator5Result}    ${voteResult5}

wait for the change of mediator
    Log    wait for 100s
    Sleep    200

choose the winner of mediator
    #票数高的应该成为活跃的mediator
    ${isActive}=    Create List
    run keyword if    ${mediator1Result}>0    Append To List    ${isActive}    ${mediatorHolder1}
    run keyword if    ${mediator2Result}>0    Append To List    ${isActive}    ${mediatorHolder2}
    run keyword if    ${mediator3Result}>0    Append To List    ${isActive}    ${mediatorHolder3}
    run keyword if    ${mediator4Result}>0    Append To List    ${isActive}    ${mediatorHolder4}
    run keyword if    ${mediator5Result}>0    Append To List    ${isActive}    ${mediatorHolder5}
    ${firstAddr}=    Get From List    ${isActive}    0
    ${secondAddr}=    Get From List    ${isActive}    1
    Set Global Variable    ${activeAccount1}    ${firstAddr}
    Set Global Variable    ${activeAccount2}    ${secondAddr}
    log    ${activeAccount1}    INFO
    log    ${activeAccount2}    INFO

determine if the change mediator
    ${params}=    Create List    ${activeAccount1}
    ${respJson}=    sendRpcPost    ${IsActiveMediator}    ${params}    JudgingResults
    ${res}    set variable    ${respJson["result"]}
    ${params1}=    Create List    ${activeAccount1}
    ${respJson1}=    sendRpcPost    ${IsActiveMediator}    ${params1}    JudgingResults
    ${res1}    set variable    ${respJson["result"]}
    Should Be Equal    ${res}    ${true}
    Should Be Equal    ${res1}    ${true}
    log    ${res}    INFO
