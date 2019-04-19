*** Settings ***
Library           RequestsLibrary
Library           Collections
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${geneAdd}        P1HhWxfQLMgb5TfE56GASURCuitX2XL397G
${recieverAdd}    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw
${contractId}     PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
${result_code}    [a-z0-9]{64}
${preTokenId}     QA001

*** Test Cases ***
transferToken_Nomal
    [Tags]    normal
    [Template]    normalTransferToken
    ${result_code}
