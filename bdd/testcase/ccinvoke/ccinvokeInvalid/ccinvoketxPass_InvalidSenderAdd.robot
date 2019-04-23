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
Scenario: invalidSenderAdd
    [Template]    InvalidCcinvoke
    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence
    ...    2    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    1    ${6000}    -32000
    ...    multiple keys match address    ${EMPTY}
    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence
    ...    2    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    1    ${6000}    -32000
    ...    multiple keys match address    P
    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence
    ...    2    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    1    ${6000}    -32000
    ...    multiple keys match address    P1
    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence
    ...    2    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    1    ${6000}    -32000
    ...    multiple keys match address    P1HhWxfQLMgb5TfE56GASURCuitX2XL397
    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence
    ...    2    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    1    ${6000}    -32000
    ...    multiple keys match address    P1HhWxfQLMgb5TfE56GASURCuitX2XL397G1
    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    200    2    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43    createToken    QA666    evidence
    ...    2    1000    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw    1    ${6000}    -32000
    ...    multiple keys match address    P1HhWxfQLMgb5TfE56GAFE8332432KK4JUWL

*** Keywords ***
