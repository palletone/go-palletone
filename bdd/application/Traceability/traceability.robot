*** Settings ***
Test Setup        beforeVote
Library           Collections
Resource          ../pubVariables.robot
Resource          ../pubFuncs.robot
Resource          ../setups.robot

*** Test Cases ***
traceability
    Given user create token
    When user create traceability
    and create proof existence
    and user tranfer token
    and user query traceability
    and user query proofexistence
    Then user query tokenhistory

*** Keywords ***
user create token
    Log    user create token
    ${args}=    Create List    ${createMethod}    ${str1}    ${symbol}    ${int2}    ${int1}
    ...    ${str2}    ${strnull}
    ${params}=    Create List    ${userAccount}    ${userAccount}    100    1    ${contractAddr2}
    ...    ${args}    ${int1}    ${120}    ${strnull}
    ${resp}=    sendRpcPost    ${createToken}    ${params}    users create token
    Sleep    5
    Log    user query token
    ${assetId}    queryToken
    ${tok}    set variable    ${assetId}-${tokenid}
    Log    ${tok}
    Set Global Variable    ${token}    ${tok}
    Sleep    5

user create traceability
    Log    user create token
    ${args}=    Create List
    ${params}=    Create List    ${userAccount2}    ${tokenid}    ${symbol}    ${maindata}    ${extradata}
    ...    ${token}
    ${resp}=    sendRpcPost    ${createTraceability}    ${params}    create token Traceability
    ${res}    set variable    ${resp["result"]}
    log    ${res}    INFO
    Sleep    10

create proof existence
    Log    create proof existence
    ${args}=    Create List
    ${params}=    Create List    ${userAccount2}    ${maindata1}    ${extradata1}    ${reference1}    ${fee}
    ${resp}=    sendRpcPost    ${createProofExistence}    ${params}    create proof existence
    ${res}    set variable    ${resp["result"]}
    log    ${res}    INFO
    Sleep    10

user tranfer token
    Log    user tranfer token
    ${args}=    Create List
    ${params}=    Create List    ${token}    ${userAccount2}    ${userAccount3}    ${int1}    ${int1}
    ...    ${strnull}    ${int1}    ${120}
    ${resp}=    sendRpcPost    ${tranferToken}    ${params}    user tranfer token
    Sleep    5

user query traceability
    Log    user query traceability
    Log    user query Asset
    ${args}=    Create List
    ${params}=    Create List    ${token}
    ${resp}=    sendRpcPost    ${queryToken}    ${params}    user query Asset
    Sleep    5
    ${queryresult}=    Get From Dictionary    ${resp}    result
    ${firstRes}=    Get From List    ${queryresult}    0
    ${main}    set variable    ${firstRes["main_data"]}
    ${extra}    set variable    ${firstRes["extra_data"]}
    ${ref}    set variable    ${firstRes["reference"]}
    Should Be Equal    ${main}    ${maindata}
    Should Be Equal    ${extra}    ${extradata}
    Should Be Equal    ${ref}    ${token}
    Log    ${main}
    Log    ${extra}
    Log    ${ref}

user query proofexistence
    Log    user query proofexistence
    ${args}=    Create List
    ${params}=    Create List    ${reference1}
    ${resp}=    sendRpcPost    ${checkProofExistence}    ${params}    check proof existence
    ${res}=    Get From Dictionary    ${resp}    result
    ${refs}=    Get From List    ${res}    0
    ${ref}    set variable    ${refs["reference"]}
    Should Be Equal    ${ref}    ${reference1}
    Log    ${ref}
    Sleep    5
    ${args}=    Create List
    ${params}=    Create List    ${maindata1}
    ${resp1}=    sendRpcPost    ${queryMaindata}    ${params}    user query maindata
    ${res}    set variable    ${resp["result"]}
    ${refs}=    Get From List    ${res}    0
    ${mian}    set variable    ${refs["main_data"]}
    ${extra}    set variable    ${refs["extra_data"]}
    ${ref}    set variable    ${refs["reference"]}
    Should Be Equal    ${mian}    ${maindata1}
    Should Be Equal    ${extra}    ${extradata1}
    Should Be Equal    ${ref}    ${reference1}
    Log    ${mian}
    Sleep    5

user query tokenhistory
    Log    user query tokehistory
    ${args}=    Create List
    ${params}=    Create List    ${userAccount2}
    ${resp}=    sendRpcPost    ${queryAddrHistory}    ${params}    check proof existence
    ${res}    set variable    ${resp["result"]}
    ${refs}=    Get From List    ${res}    3
    run keyword if    ${refs["tx_hash"]}!=${args}    log    success
    Log    ${refs["tx_hash"]}
    Sleep    5
    ${args}=    Create List
    ${params}=    Create List    ${token}
    ${resp}=    sendRpcPost    ${queryTokenHistory}    ${params}    check proof existence
    ${res}    set variable    ${resp["result"]}
    ${refs}=    Get From List    ${res}    1
    run keyword if    ${refs["tx_hash"]}!=${args}    log    success
    Log    ${refs["tx_hash"]}
