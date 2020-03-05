*** Variables ***
${errorMessage}    ${EMPTY}

*** Test Cases ***
success
    ${one}    Set Variable    1
    ${two}    Set Variable    1
    Should Be Equal    ${one}    ${two}

fail
    Should Be Empty    ${errorMessage}
