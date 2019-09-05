*** Test Cases ***
success
    ${one}    Set Variable    1
    ${two}    Set Variable    1
    Should Be Equal    ${one}    ${two}

fail
    ${one}    Set Variable    1
    ${two}    Set Variable    2
    Should Be Equal    ${one}    ${two}
