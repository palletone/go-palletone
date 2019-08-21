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
    [Arguments]     ${isMultipleNodes}=${true}
    ${args}=    Create List
    ${params}=    Create List
    ${respJson}=    sendRpcPost    ${host}    ${personalListAccountsMethod}    ${params}    queryTokenHolder
    Dictionary Should Contain Key    ${respJson}    result
    ${accounts}=    Get From Dictionary    ${respJson}    result
    ${second}=    Set Variable If    ${nodenum}>1      ${1}
    ${index}=    Set Variable    ${0}
    ${index}=    Set Variable If    ${isMultipleNodes}==${true}    ${second}
    ...    ${isMultipleNodes}==${false}    ${0}
    ${addr}=    Get From List    ${accounts}    ${index}
    Set Global Variable    ${tokenHolder}    ${addr}
