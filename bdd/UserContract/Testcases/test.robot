*** Test Cases ***
test
    ${a}=    Set Variable    ${3}
    ${b}=    Set Variable    ${6}
    Run Keyword If    ${b}-${a}>3    Log    "11111"
    ...    ELSE    Fail    "value error"
