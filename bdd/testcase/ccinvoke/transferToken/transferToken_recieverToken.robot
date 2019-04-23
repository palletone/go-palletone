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
${preTokenId}     QA010
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
    normalCcinvokePass    ${result_code}    ${preTokenId}    ${tokenDecimal}    ${tokenAmount}    ${amount}    ${poundage}
    sleep    3
    ${resultGene}    getBalance    ${GeneAdd}
    ${result1}    getBalance    ${recieverAdd}
    ${key}    getTokenId    ${preTokenId}    ${resultGene}
    ${item}    Set Variable    0
    #${item}    Get From Dictionary    ${result1}    ${key}
    #${item1}    Run Keyword If    ${item}==''    Set Variable    0
    #ELSE    Get From Dictionary    ${result1}    ${key}
    ${tokenResult}    transferToken    ${key}    ${GeneAdd}    ${recieverAdd}    ${senderAmount}    ${pdg}
    ...    ${evidence}    ${unlocktime}
    sleep    2
    ${item1}    Evaluate    ${item}+${senderAmount}
    ${result2}    getBalance    ${recieverAdd}
    ${key2}    getTokenId    ${preTokenId}    ${result2}
    ${item2}    Get From Dictionary    ${result2}    ${key2}
    Should Be Equal As Numbers    ${item2}    ${item1}
