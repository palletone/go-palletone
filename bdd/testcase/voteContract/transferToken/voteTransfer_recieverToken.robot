*** Settings ***
Default Tags      normal
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***

*** Test Cases ***
Scenario: Vote Contract - Transfer Token
    [Documentation]    Verify Reciever's Transfer PTN
    ${geneAdd}    Given Get genesis address
    #${ret}    When Create token of vote contract    ${geneAdd}
    ${key}    ${item1}    And Request getbalance before create token    ${geneAdd}
    And Request transfer token
    ${item1}    And Calculate gain of recieverAdd    ${item1}
    ${item2}    And Request getbalance after create token    ${key}
    Then Assert gain of reciever    ${item1}    ${item2}

*** Keywords ***
Get genesis address
    ${geneAdd}    getGeneAdd    ${host}
    sleep    4
    [Return]    ${geneAdd}

Request getbalance before create token
    [Arguments]    ${geneAdd}
    ${PTN1}    ${result1}    normalGetBalance    ${geneAdd}
    sleep    2
    ${key}    getTokenId    ${voteId}    ${result1['result']}
    sleep    2
    ${PTN2}    ${result2}    normalGetBalance    ${recieverAdd}
    sleep    3
    #${dicRes}    Evaluate    demjson.encode(${result2})    demjson
    #log    type(${dicRes})
    #${jsonRes}    To Json    ${dicRes}
    #: FOR    ${keys}    IN    ${dicRes}
    #\    log    ${keys}
    #${strResult}    Evaluate    str(${jsonRes})
    ${item1}    voteExist    ${key}    ${result2}
    sleep    2
    [Return]    ${key}    ${item1}

Request transfer token
    ${tokenResult}    transferToken    ${key}    ${geneAdd}    ${recieverAdd}    ${PTNAmount}    ${PTNPoundage}
    ...    ${evidence}    ${duration}
    sleep    5

Calculate gain of recieverAdd
    [Arguments]    ${item1}
    ${item1}    Evaluate    ${item1}+${PTNAmount}
    [Return]    ${item1}

Request getbalance after create token
    [Arguments]    ${key}
    ${result2}    getBalance    ${recieverAdd}
    sleep    2
    ${item2}    Get From Dictionary    ${result2}    ${key}
    sleep    3
    [Return]    ${item2}

Assert gain of reciever
    [Arguments]    ${item1}    ${item2}
    Should Be Equal As Strings    ${item2}    ${item1}
