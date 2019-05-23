*** Test Cases ***
aliceUsePRC720Token
    Given Alice issues her personal token named ALICE, amount is 10^9, dicimal is 10^8 succeed
    When Alice transfers 5*10^8 ALICE to Bob succeed
    And Alice increases ALICE 10*9 succeed
    And Alice transfers 5*10^8 ALICE to Carol succeed
    Then Alice has 10^9 ALICE left

*** Keywords ***
Alice issues her personal token named ALICE, amount is 10^9, dicimal is 10^8 succeed
    Log    "1"

Alice transfers 5*10^8 ALICE to Bob succeed
    Log    "2"

Alice increases ALICE 10*9 succeed
    Log    "3"

Alice transfers 5*10^8 ALICE to Carol succeed
    Log    "4"

Alice has 10^9 ALICE left
    Log    "5"
