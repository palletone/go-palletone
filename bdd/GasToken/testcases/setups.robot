*** Settings ***
Resource          pubVariables.robot
Resource          pubFuncs.robot
Library           Collections

*** Keywords ***
newAccounts
    # create power intermediate certificate holder
    ${user1}=    newAccount
    Set Global Variable    ${Alice}    ${user1}
    # create section intermediate certificate holder
    ${user2}=    newAccount
    Set Global Variable    ${Bob}    ${user2}
    # create user certificate holder
    ${user3}=    newAccount
    Set Global Variable    ${Carol}    ${user3}

queryTokenHolder
    ${args}=    Create List
    ${params}=    Create List
    ${respJson}=    sendRpcPost    ${personalListAccountsMethod}    ${params}    queryTokenHolder
    Dictionary Should Contain Key    ${respJson}    result
    ${accounts}=    Get From Dictionary    ${respJson}    result
    ${firstAddr}=    Get From List    ${accounts}    0
    Set Global Variable    ${tokenHolder}    ${firstAddr}
