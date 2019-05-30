*** Settings ***
Default Tags      normal
Library           RequestsLibrary
Library           Collections
Library           /usr/lib/python2.7/decimal.py
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${geneAdd}        P1CwGYGSjWSaJrysHAjAWtDyFSsbcYwoULv
${recieverAdd}    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw
${contractId}     PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
${result_code}    [a-z0-9]{64}
${tokenId}        QA001
${tokenDecimal}    1
${tokenAmount}    25000
${amount}         2000
${poundage}       1

*** Test Cases ***
Ccinvoke SenderPTN
    ${log}    getGeneAdd    ${host}
    ${PTN1}    ${result}    normalGetBalance    ${log}
    normalCcinvokePass    ${result_code}    ${tokenId}    ${tokenDecimal}    ${tokenAmount}    ${amount}    ${poundage}
    sleep    3
    ${GAIN}    countRecieverPTN    2001
    ${log2}    getGeneAdd    ${host}
    ${PTN2}    ${result}    normalGetBalance    ${log2}
    sleep    4
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')-decimal.Decimal('${GAIN}')    decimal
    Should Be Equal As Numbers    ${PTN2}    ${PTNGAIN}
