*** Variables ***
${errorMessage}    ${EMPTY}

*** Test Cases ***
success
    ${one}    Set Variable    1
    ${two}    Set Variable    1
    Should Be Equal    ${one}    ${two}

fail
    log    ${errorMessage}
    ${one}    Set Variable    1
    ${two}    Set Variable    2
    Should Be Equal    ${one}    ${two}
