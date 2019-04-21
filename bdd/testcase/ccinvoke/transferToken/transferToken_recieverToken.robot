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
${tokenId}        QA004
${tokenDecimal}    1
${tokenAmount}    25000
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
    normalCcinvokePass    ${result_code}    ${tokenId}    ${tokenDecimal}    ${tokenAmount}    ${amount}    ${poundage}
    sleep    2
    ${resultGe}    getBalance    ${GeneAdd}
    ${keyGe}    ${itemGe}    getTokenStarts    ${tokenId}    ${resultGe}
    ${result1}    getBalance    ${recieverAdd}
    ${key}    ${item1}    getTokenStarts    ${tokenId}    ${result1}
    #${item1}    Run Keyword If    ${strResult}=={}    Set Variable    ${0}
    ${tokenResult}    transferToken    ${keyGe}    ${GeneAdd}    ${recieverAdd}    ${senderAmount}    ${pdg}
    ...    ${evidence}    ${unlocktime}
    ${item1}    Evaluate    ${item1}+${senderAmount}
    sleep    3
    ${result2}    getBalance    ${recieverAdd}
    ${key2}    ${item2}    getTokenStarts    ${tokenId}    ${result2}
    Should Be Equal As Numbers    ${item2}    ${item1}
