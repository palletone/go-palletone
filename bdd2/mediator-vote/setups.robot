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
    ${user4}=    newAccount
    Set Global Variable    ${userAccount4}    ${user4}
    ${user5}=    newAccount
    Set Global Variable    ${userAccount5}    ${user5}

transferPTNToUser
    # transfer PTN to    user for the poll
    transferPTN    ${userAccount}
    Log    wait for tx being packaged into unit
    Sleep    5    # should sleep, because transaction has not been packaged into unit
    ${balance}=    getBalance    ${userAccount}
    Should Be Equal    ${balance}    ${amount}
    transferPTN2    ${userAccount2}
    Log    wait for tx being packaged into unit
    Sleep    5    # should sleep, because transaction has not been packaged into unit
    ${balance2}=    getBalance    ${userAccount2}
    Should Be Equal    ${balance2}    ${amount2}
    transferPTN3    ${userAccount3}
    Log    wait for tx being packaged into unit
    Sleep    5    # should sleep, because transaction has not been packaged into unit
    ${balance3}=    getBalance    ${userAccount3}
    Should Be Equal    ${balance3}    ${amount3}
    transferPTN4    ${userAccount4}
    Log    wait for tx being packaged into unit
    Sleep    5    # should sleep, because transaction has not been packaged into unit
    ${balance4}=    getBalance    ${userAccount4}
    Should Be Equal    ${balance4}    ${amount4}
    transferPTN5    ${userAccount5}
    Log    wait for tx being packaged into unit
    Sleep    5    # should sleep, because transaction has not been packaged into unit
    ${balance5}=    getBalance    ${userAccount5}
    Should Be Equal    ${balance5}    ${amount5}
