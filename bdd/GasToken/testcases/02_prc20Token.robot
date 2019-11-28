*** Settings ***
Resource          pubVariables.robot
Resource          pubFuncs.robot
Library           Collections
Library           BuiltIn

*** Test Cases ***
aliceUsePRC20Token
    Given Alice issues her personal token named ALICE, amount is 1000, decimal is 1 succeed
    When Alice transfers 500 ALICE to Bob succeed
    And Bob transfers 200 ALICE to Carol succeed
    And Alice increases 1000 ALICE succeed
    And Alice transfers 500 ALICE to Carol succeed
    Then Alice has 1000 ALICE left
    And Bob has 300 ALICE left
    And Carol has 700 ALICE left

*** Keywords ***
Alice issues her personal token named ALICE, amount is 1000, decimal is 1 succeed
    unlockAccount    ${Alice}
    issueToken    ${Alice}    ${AliceToken}    1000    1    Alice's token
    Wait for transaction being packaged
    ${balance}=    getAllBalance    ${Alice}
    ${tokenIDs}=    Get Dictionary Keys    ${balance}
    : FOR    ${id}    IN    @{tokenIDs}
    \    Continue For Loop if    '${id}'=='${gasToken}'
    \    Set Global Variable    ${AliceTokenID}    ${id}

Alice transfers 500 ALICE to Bob succeed
    transferTokenTo    ${AliceTokenID}    ${Alice}    ${Bob}    500    1
    Wait for transaction being packaged

Bob transfers 200 ALICE to Carol succeed
    transferTokenTo    ${AliceTokenID}    ${Bob}    ${Carol}    200    1
    Wait for transaction being packaged

Alice increases 1000 ALICE succeed
    supplyToken    ${Alice}    ${AliceToken}    1000
    Wait for transaction being packaged

Alice transfers 500 ALICE to Carol succeed
    ${respJson}=    transferTokenTo    ${AliceTokenID}    ${Alice}    ${Carol}    500    1
    Wait for transaction being packaged

Alice has 1000 ALICE left
    ${balance}=    getAllBalance    ${Alice}
    ${amount}=    Get From Dictionary    ${balance}    ${AliceTokenID}
    Should Be Equal    ${amount}    1000

Bob has 300 ALICE left
    ${balance}=    getAllBalance    ${Bob}
    ${amount}=    Get From Dictionary    ${balance}    ${AliceTokenID}
    Should Be Equal    ${amount}    300

Carol has 700 ALICE left
    ${balance}=    getAllBalance    ${Carol}
    ${amount}=    Get From Dictionary    ${balance}    ${AliceTokenID}
    Should Be Equal    ${amount}    700
