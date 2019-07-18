*** Settings ***
Force Tags        invalidSign
Default Tags      invalidSign
Library           RequestsLibrary
Library           Collections
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/invalidKwd.txt
Resource          ../../utilKwd/behaveKwd.txt
Resource          ../../utilKwd/normalKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${method}         wallet_signRawTransaction
${geneAdd}        P1CwGYGSjWSaJrysHAjAWtDyFSsbcYwoULv
${recieverAdd}    P11LKXXsDuKUdo3cJEy7uMnqnvw4kwHPs8
${result_code}    \f[a-z0-9]*

*** Test Cases ***
signPassInvalid1
    [Tags]    invalidSign1
    ${crtResult}    normalCrtTrans    ${result_code}
    InvalidSignTrans    ${crtResult}    ALL    ${EMPTY}    -32000    get addr by outpoint get err:could not decrypt key with given passphrase

signPassInvalid2
    [Tags]    invalidSign1
    ${crtResult}    normalCrtTrans    ${result_code}
    InvalidSignTrans    ${crtResult}    ALL    2    -32000    could not decrypt key with given passphrase
    InvalidSignTrans    ${crtResult}    ALL    1sf    -32000    could not decrypt key with given passphrase

signPassInvalid3
    [Tags]    invalidSign1
    ${crtResult}    normalCrtTrans    ${result_code}
    InvalidSignTrans    ${crtResult}    ALL    @    -32000    get addr by outpoint get err:could not decrypt key with given passphrase
    InvalidSignTrans    ${crtResult}    ALL    FDEW    -32000    get addr by outpoint get err:could not decrypt key with given passphrase

*** Keywords ***
