*** Settings ***
Library           RequestsLibrary
Library           Collections
Resource          pubVariables.robot
Resource          pubFuncs.robot
Resource          setups.robot

*** Test Cases ***
testQuery
    ${respJson}=    queryCert    P1M1fT4E86Agm8HN7iFrdrr4uzgJBUz4RFX
    Dictionary Should Contain Key    ${respJson}    result

testAddCert
    ${respJson}=    addCert    P1M1fT4E86Agm8HN7iFrdrr4uzgJBUz4RFX    addServerCert    P1Vuup9iys7mgFURQhMfa8g8Xipv7vBW9e    certbytes
    Dictionary Should Not Contain Key    ${respJson}    error

logCertHolders
    Log    ${intermediate1CertHolder}
    Log    ${intermediate2CertHolder}
    Log    ${userCertHolder}

testSendPTN
    [Setup]    beforeIssueCerts
    transferPTN    ${intermediate1CertHolder}
    Sleep    5    # should sleep, because transaction has not been packaged into unit
    ${balance}=    getBalance    ${intermediate1CertHolder}
    Should Be Equal    ${balance}    ${amount}
