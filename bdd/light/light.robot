*** Settings ***
Library           RequestsLibrary
Library           Collections

*** Variables ***
${syncutxoinfo1}    ${EMPTY}
${syncutxoinfo2}    ${EMPTY}
${syncutxoinfo3}    ${EMPTY}
${balance1}       ${EMPTY}
${balance2}       ${EMPTY}
${balance3}       ${EMPTY}

*** Test Cases ***
ligth
    log    ${syncutxoinfo1}
    Should Be Equal As Strings    ${syncutxoinfo1}    "ok"
    log    ${balance1}
    Should Be Equal As Strings    ${balance1}    100
    log    ${syncutxoinfo2}
    Should Be Equal As Strings    ${syncutxoinfo2}    "ok"
    log    ${balance2}
    Should Be Equal As Strings    ${balance2}    80
    log    ${syncutxoinfo3}
    Should Be Equal As Strings    ${syncutxoinfo3}    "ok"
    log    ${balance3}
    Should Be Equal As Strings    ${balance3}    50
