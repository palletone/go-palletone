*** Settings ***
Library           RequestsLibrary
Library           Collections
Library           ../../utilFunc/createToken.py
Resource          ../../utilKwd/normalKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt

*** Variables ***
${host}           http://localhost:8545/
${geneAdd}        P17XYSQ4qBKeWF9qicEdG5ZzfvTZQke4Ys9
${recieverAdd}    P1MdMxNVaKZYdBBFB8Fszt8Bki1AEmRRSxw
${contractId}     PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
${tokenId}        QA001
${tokenDecimal}    1
${tokenAmount}    2500
${amount}         2000
${poundage}       1
${senderAmount}    2
${pdg}            1
${evidence}       evidence
${unlocktime}     ${6000000}
${result_code}    [a-z0-9]{64}

*** Test Cases ***
transferToken_recieverToken
    [Tags]    normal
    ${GeneAdd}    getGeneAdd    ${host}
    #normalCcinvokePass    ${result_code}    ${tokenId}    ${tokenDecimal}    ${tokenAmount}    ${amount}    ${poundage}
    #sleep    2
    ${PTN1}    ${result1}    normalGetBalance    ${recieverAdd}
    ${strResult}    Evaluate    str(${result1['result']})
    ${key}    getTokenId    ${tokenId}    ${result1['result']}
    ${item}    Get From Dictionary    ${result1['result']}    ${key}
    ${item1}    Run Keyword If    ${strResult}=={}    Set Variable    ${0}
    ...    ELSE    Set Variable    ${item}
    ${tokenResult}    transferToken    ${key}    ${GeneAdd}    ${recieverAdd}    ${senderAmount}    ${pdg}
    ...    ${evidence}    ${unlocktime}
    sleep    2
    ${item1}    Evaluate    ${item1}+${senderAmount}
    ${PTN2}    ${result2}    normalGetBalance    ${recieverAdd}
    ${key}    getTokenId    ${tokenId}    ${result2['result']}
    ${item2}    Get From Dictionary    ${result2['result']}    ${key}
    Should Be Equal As Numbers    ${item2}    ${item1}
