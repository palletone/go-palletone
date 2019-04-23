*** Settings ***
Force Tags        invalidAdd
Default Tags      invalidAdd
Library           RequestsLibrary
Library           Collections
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/invalidKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt
Resource          ../../utilKwd/normalKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${method}         ptn_ccinvoketxPass

*** Test Cases ***
Scenario: invalidContractId
    [Template]    InvalidCcinvoke
    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    200    2    {EMPTY}    createToken    QA666    evidence
    ...    2    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    1    ${6000}    -32000
    ...    ContractInvokeReq request param is error
    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    200    2    $    createToken    QA666    evidence
    ...    2    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    1    ${6000}    -32000
    ...    ContractInvokeReq request param is error
    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG4    createToken    QA666    evidence
    ...    2    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    1    ${6000}    -32000
    ...    ContractInvokeReq request param is error
    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG433    createToken    QA666    evidence
    ...    2    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    1    ${6000}    -32000
    ...    ContractInvokeReq request param is error
    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG54    createToken    QA666    evidence
    ...    2    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    1    ${6000}    -32000
    ...    ContractInvokeReq request param is error

*** Keywords ***
