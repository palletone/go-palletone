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

transferPTNToUser
    # transfer PTN to    user for the poll
    transferPTN    ${userAccount}
    Log    wait for tx being packaged into unit
    Sleep    5    # should sleep, because transaction has not been packaged into unit
    ${balance}=    getBalance    ${userAccount}
    Should Be Equal    ${balance}    ${amount}
