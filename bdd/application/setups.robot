*** Settings ***
Resource          pubVariables.robot
Resource          pubFuncs.robot
Library           Collections

*** Keywords ***
beforeVote
    queryTokenHolder
    newAccounts
    transferPTNToUser

newAccounts
    # create a user for the poll
    ${user1}=    newAccount
    Set Global Variable    ${userAccount}    ${user1}
    ${user2}=    newAccount
    Set Global Variable    ${userAccount2}    ${user2}
    ${user3}=    newAccount
    Set Global Variable    ${userAccount3}    ${user3}

transferPTNToUser
    # transfer PTN to    user for the poll
    transferPTN    ${userAccount}
    transferPTN    ${userAccount2}
    transferPTN    ${userAccount3}
    Log    wait for tx being packaged into unit
    Sleep    5    # should sleep, because transaction has not been packaged into unit
    ${balance}=    getBalance    ${userAccount}
    Should Be Equal    ${balance}    ${amount}
    ${balance}=    getBalance    ${userAccount2}
    Should Be Equal    ${balance}    ${amount}
    ${balance}=    getBalance    ${userAccount3}
    Should Be Equal    ${balance}    ${amount}
