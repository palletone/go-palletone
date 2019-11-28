*** Settings ***
Default Tags      invalidAdd
Resource          ../../utilKwd/utilVariables.txt
Resource          ../../utilKwd/invalidKwd.txt
Resource          ../../utilKwd/utilDefined.txt
Resource          ../../utilKwd/behaveKwd.txt
Resource          ../../utilKwd/normalKwd.txt

*** Variables ***
${host}           http://localhost:8545/

*** Test Cases ***
senderInvalid1
    [Tags]    invalidAdd2
    [Template]    InvalidCrtTrans
    P1    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    2    -32000    sender address is invalid

senderInvalid2
    [Tags]    invalidAdd2
    [Template]    InvalidCrtTrans
    P    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    2    -32000    sender address is invalid

senderInvalid3
    [Tags]    invalidAdd2
    [Template]    InvalidCrtTrans
    f    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    2    -32000    sender address is invalid

senderEmpty
    [Tags]    invalidAdd2
    [Template]    InvalidCrtTrans
    ${Empty}    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    2    -32000    sender address is empty

recieverInvalid1
    [Tags]    invalidAdd1
    [Template]    InvalidCrtTrans
    P1FRZ2AVgCd2TwS5SYDy1ehe8YaXYn86J7U    P1    10    2    -32000    receiver address is invalid

recieverInvalid2
    [Tags]    invalidAdd2
    [Template]    InvalidCrtTrans
    P1FRZ2AVgCd2TwS5SYDy1ehe8YaXYn86J7U    P    10    2    -32000    receiver address is invalid

recieverInvalid3
    [Tags]    invalidAdd2
    [Template]    InvalidCrtTrans
    P1FRZ2AVgCd2TwS5SYDy1ehe8YaXYn86J7U    f    10    2    -32000    receiver address is invalid

recieverEmpty
    [Tags]    invalidAdd1
    [Template]    InvalidCrtTrans
    P1FRZ2AVgCd2TwS5SYDy1ehe8YaXYn86J7U    ${Empty}    10    2    -32000    receiver address is empty

amountInvalid1
    [Tags]    invalidAmount
    [Template]    InvalidCrtTrans
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    ferew    2    -32602    invalid argument 2: Error decoding string

amountInvalid2
    [Tags]    invalidAmount
    [Template]    InvalidCrtTrans
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    3d    2    -32602    invalid argument 2: Error decoding string

amountInvalid3
    [Tags]    invalidAmount
    [Template]    InvalidCrtTrans
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    $    2    -32602    invalid argument 2: Error decoding string

amountInvalid4
    [Tags]    invalidAmount
    [Template]    InvalidCrtTrans
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    0    2    -32000    amounts is invalid

amountInvalid5
    [Tags]    invalidAmount
    [Template]    InvalidCrtTrans
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    0    2    -32000    amounts is invalid
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    -6    2    -32000    amounts is invalid

amountEmpty
    [Tags]    invalidAmount
    [Template]    InvalidCrtTrans
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    ${Empty}    2    -32602    invalid argument 2: Error decoding string

poundageInvalid1
    [Tags]    invalidAmount
    [Template]    InvalidCrtTrans
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    r443    -32602    invalid argument 3: Error decoding string
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    4fgf    -32602    invalid argument 3: Error decoding string
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    ${SPACE}    -32602    invalid argument 3: Error decoding string
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    $    -32602    invalid argument 3: Error decoding string
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    "0"    -32602    invalid argument 3: Error decoding string

poundageInvalid2
    [Tags]    invalidAmount
    [Template]    InvalidCrtTrans
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    0    -32000    fee is invalid
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    -4.5    -32000    fee is invalid

poundageInvalid3
    [Tags]    invalidAmount
    [Template]    InvalidCrtTrans
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    10000000000001    -32000    Select utxo err

poundageEmpty
    [Tags]    invalidAmount
    [Template]    InvalidCrtTrans
    P122EGDTLmfaMCF5YTkre8Zd9urLV2y2coy    P1MhaR76qdVPJMJhUYMhfzdEoVndvmEWMUX    10    ${Empty}    -32602    invalid argument 3: Error decoding string

*** Keywords ***
