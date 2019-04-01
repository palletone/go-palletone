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
${method}         ptn_signRawTransaction
#${crtResult}     f8bff8bdf8bb80b8b8f8b6e6e58080a04d5ad1c713f9386488e6ec5c4c5c7bf43a4eb29114a6b2ac8754e181da4702898080f88cf842843b9aca009976a91488be35cc7c2cace89a62193d96cb45e93d808ec688ace290400082bb0800000000000000000000009000000000000000000000000000000000f84688016345781bf955009976a91482ecd121e7ba7aca3d27b6d86d41aafa0b59f84988ace290400082bb080000000000000000000000900000000000000000000000000000000080
${geneAdd}        P1CwGYGSjWSaJrysHAjAWtDyFSsbcYwoULv
${recieverAdd}    P11LKXXsDuKUdo3cJEy7uMnqnvw4kwHPs8
@{transList}      ${geneAdd}    ${recieverAdd}    10    2
${result_code}    \f[a-z0-9]*

*** Test Cases ***
signPassInvalid1
    [Tags]    invalidSign1
    [Template]    setCrtInvalidSign
    2    -32000    could not decrypt key with given passphrase
    1sf    -32000    could not decrypt key with given passphrase
    ${Empty}    -32000    could not decrypt key with given passphrase

signPassInvalid2
    [Tags]    invalidSign1
    [Template]    setInvalidSignPass
    \    1    -32000    Params is empty

signPassInvalid3
    [Tags]    invalidSign1
    [Template]    setInvalidSignPass
    1    1    -32000    Params decode is invalid

signPassInvalid4
    [Tags]    invalidSign1
    [Template]    setInvalidSignPass
    @    1    -32000    Params is invalid
    FDEW    1    -32000    Params is invalid

*** Keywords ***
I set signTrans password to ${i}
    [Arguments]    ${crtResult}    ${i}
    @{transList}    Create List    ${crtResult}    ${i}
    [Return]    @{transList}
