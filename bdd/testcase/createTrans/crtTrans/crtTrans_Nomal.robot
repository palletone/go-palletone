*** Settings ***
Library           RequestsLibrary
Library           Collections
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${geneAdd}        P1CwGYGSjWSaJrysHAjAWtDyFSsbcYwoULv
${recieverAdd}    P11LKXXsDuKUdo3cJEy7uMnqnvw4kwHPs8
@{transList}      ${geneAdd}    ${recieverAdd}    10    2
${result_code}    \f[a-z0-9]*

*** Test Cases ***
crtTrans normal
    [Tags]    normal
    [Template]    normalCrtTrans
    ${result_code}
