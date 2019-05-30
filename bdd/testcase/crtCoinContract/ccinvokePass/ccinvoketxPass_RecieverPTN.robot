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
${tokenId}        QA003
${tokenDecimal}    1
${tokenAmount}    25000
${poundage}       1
${amount}         2000
${gain}           2000

*** Test Cases ***
Ccinvoke RecieverPTN
    ${PTN1}    ${result}    normalGetBalance    ${recieverAdd}
    normalCcinvokePass    ${result_code}    ${tokenId}    ${tokenDecimal}    ${tokenAmount}    ${amount}    1
    sleep    2
    ${gain1}    countRecieverPTN    ${gain}
    ${PTNGAIN}    Evaluate    decimal.Decimal('${PTN1}')+decimal.Decimal('${gain1}')    decimal
    sleep    2
    ${PTN2}    ${result}    normalGetBalance    ${recieverAdd}
    sleep    4
    Should Be Equal As Numbers    ${PTN2}    ${PTNGAIN}
