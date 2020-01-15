*** Settings ***
Test Setup        beforeVote
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot
Resource          ../setups.robot

*** Test Cases ***
mediatorvote
    Given user unlock its account succeed
    When pledge users ptn
    and check mediator list
    and wait for transaction being packaged
    and check mediator list
    and check mediator actives list
    Then vote mediator account

*** Keywords ***
user unlock its account succeed
    Log    " user unlock its account succeed"
    ${respJson}=    unlockAccount    ${userAccount}
    ${respJson2}=    unlockAccount    ${userAccount2}
    ${respJson3}=    unlockAccount    ${userAccount3}
    ${respJson4}=    unlockAccount    ${userAccount4}
    ${respJson5}=    unlockAccount    ${userAccount5}
    log    ${respJson}
    Dictionary Should Contain Key    ${respJson}    result
    Should Be Equal    ${respJson["result"]}    ${true}

pledge users ptn
    Log    " pledge users ptn succeed"
    ${args}=    Create List    ${pledgeMethod}
    ${params}=    Create List    ${userAccount}    ${contractAddr}    1000    1    ${contractAddr}
    ...    ${args}    ${strnull}    ${strnull}
    ${params2}=    Create List    ${userAccount2}    ${contractAddr}    2000    1    ${contractAddr}
    ...    ${args}    ${strnull}    ${strnull}
    ${params3}=    Create List    ${userAccount3}    ${contractAddr}    3000    1    ${contractAddr}
    ...    ${args}    ${strnull}    ${strnull}
    ${params4}=    Create List    ${userAccount4}    ${contractAddr}    4000    1    ${contractAddr}
    ...    ${args}    ${strnull}    ${strnull}
    ${params5}=    Create List    ${userAccount5}    ${contractAddr}    5000    1    ${contractAddr}
    ...    ${args}    ${strnull}    ${strnull}
    ${resp}=    sendRpcPost    ${pledgeDeposit}    ${params}    pledge users ptn
    ${resp}=    sendRpcPost    ${pledgeDeposit}    ${params2}    pledge users ptn
    ${resp}=    sendRpcPost    ${pledgeDeposit}    ${params3}    pledge users ptn
    ${resp}=    sendRpcPost    ${pledgeDeposit}    ${params4}    pledge users ptn
    ${resp}=    sendRpcPost    ${pledgeDeposit}    ${params5}    pledge users ptn

check mediator list
    Log    "check mediator list succeed"
    ${args}=    Create List
    ${params}=    Create List
    ${resp}=    sendRpcPost    ${checkMediatorList}    ${params}    check mediator list
    ${mediatorAddrs}=    Get From Dictionary    ${resp}    result
    ${firstAddr}=    Get From List    ${mediatorAddrs}    0
    ${secondAddr}=    Get From List    ${mediatorAddrs}    1
    ${thirdAddr}=    Get From List    ${mediatorAddrs}    2
    ${fourthAddr}=    Get From List    ${mediatorAddrs}    3
    ${fifthAddr}=    Get From List    ${mediatorAddrs}    4
    Set Global Variable    ${mediatorHolder1}    ${firstAddr}
    Set Global Variable    ${mediatorHolder2}    ${secondAddr}
    Set Global Variable    ${mediatorHolder3}    ${thirdAddr}
    Set Global Variable    ${mediatorHolder4}    ${fourthAddr}
    Set Global Variable    ${mediatorHolder5}    ${fifthAddr}

check mediator actives list
    Log    "check mediator list succeed"
    ${args}=    Create List
    ${params}=    Create List
    ${resp}=    sendRpcPost    ${viewMediatorActives}    ${params}    view mediator actives
    ${mediatorAddrs}=    Get From Dictionary    ${resp}    result
    ${firstAddr}=    Get From List    ${mediatorAddrs}    0
    ${secondAddr}=    Get From List    ${mediatorAddrs}    1
    ${thirdAddr}=    Get From List    ${mediatorAddrs}    2
    Set Global Variable    ${mediatorActives1}    ${firstAddr}
    Set Global Variable    ${mediatorActives2}    ${secondAddr}
    Set Global Variable    ${mediatorActives3}    ${thirdAddr}

vote mediator account
    ${args1}    run keyword if    '${mediatorHolder1}'!='${mediatorActives1}'and'${mediatorHolder1}'!='${mediatorActives2}'and'${mediatorHolder1}'!='${mediatorActives3}'    Create List    ${mediatorHolder1}
    ${params1}=    Create List    ${userAccount}    ${args1}
    ${resp}=    sendRpcPost    ${voteMediator}    ${params1}    vote mediator
    ${args2}    run keyword if    '${mediatorHolder2}'!='${mediatorActives1}'and'${mediatorHolder2}'!='${mediatorActives2}'and'${mediatorHolder2}'!='${mediatorActives3}'    Create List    ${mediatorHolder2}
    ${params2}=    Create List    ${userAccount2}    ${args2}
    ${resp}=    sendRpcPost    ${voteMediator}    ${params2}    vote mediator
    ${args3}    run keyword if    '${mediatorHolder3}'!='${mediatorActives1}'and'${mediatorHolder3}'!='${mediatorActives2}'and'${mediatorHolder3}'!='${mediatorActives3}'    Create List    ${mediatorHolder3}
    ${params3}=    Create List    ${userAccount3}    ${args3}
    ${resp}=    sendRpcPost    ${voteMediator}    ${params3}    vote mediator
    ${args4}    run keyword if    '${mediatorHolder4}'!='${mediatorActives1}'and'${mediatorHolder4}'!='${mediatorActives2}'and'${mediatorHolder4}'!='${mediatorActives3}'    Create List    ${mediatorHolder4}
    ${params4}=    Create List    ${userAccount4}    ${args4}
    ${resp}=    sendRpcPost    ${voteMediator}    ${params4}    vote mediator
    ${args5}    run keyword if    '${mediatorHolder5}'!='${mediatorActives1}'and'${mediatorHolder5}'!='${mediatorActives2}'and'${mediatorHolder5}'!='${mediatorActives3}'    Create List    ${mediatorHolder5}
    ${params5}=    Create List    ${userAccount5}    ${args5}
    ${resp}=    sendRpcPost    ${voteMediator}    ${params5}    vote mediator
    log    ${resp}
