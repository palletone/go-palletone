*** Settings ***
Resource          ../../commonlib/pubVariables.robot
Resource          ../../commonlib/pubFuncs.robot
Resource          ../../commonlib/setups.robot
Library           BuiltIn
Library           Collections

*** Test Cases ***
testprepare
    queryTokenHolder
    startProduce

*** Keywords ***
startProduce
    ${port}    set variable    ${8645}
    ${hosts}    Create List
    : FOR    ${n}    IN RANGE    ${nodenum}
    \    Run Keyword If    ${n}==${0}    startNodeProduce    ${host}
    \    Continue For Loop If    ${n}==${0}
    \    ${newport}=    Evaluate    ${port}+10*(${n}+1)
    \    ${url}=    Catenate    SEPARATOR=    http://${ip}:    ${newport}    /
    \    Append To List    ${hosts}    ${url}
    \    startNodeProduce    ${url}
    Set Global Variable    ${juryHosts}    ${hosts}
