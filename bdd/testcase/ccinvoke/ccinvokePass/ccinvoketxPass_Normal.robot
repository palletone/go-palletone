*** Settings ***
Library           RequestsLibrary
Library           Collections
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${geneAdd}        P1HhWxfQLMgb5TfE56GASURCuitX2XL397G
${recieverAdd}    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw
${contractId}     PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
${result_code}    [a-z0-9]{64}
${tokenId}        QA001
${tokenDecimal}    1
${tokenAmount}    25000
${amount}         2000
${poundage}       1

*** Test Cases ***
CcinvokePass normal
    [Template]    normalCcinvokePass
    ${result_code}    ${tokenId}    ${tokenDecimal}    ${tokenAmount}    ${amount}    ${poundage}
