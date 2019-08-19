*** Settings ***
Library           RequestsLibrary

*** Variables ***
#${ip}            123.126.106.82
#${host}          http://${ip}:58645/
${ip}             127.0.0.1
${host}           http://${ip}:8645/
${juryHosts}      Create List
${gastokenHost}    http://localhost:8545
${nodenum}        3
# methods
${ccinvokeMethod}    contract_ccinvoketx
${ccqueryMethod}    contract_ccquery
${ccinstallMethod}    contract_ccinstalltx
${ccdeployMethod}    contract_ccdeploytx
${ccstopMethod}    contract_ccstoptx
${transferPTNMethod}    wallet_transferPtn
${transferTokenMethod}    wallet_transferToken
${getBalanceMethod}    wallet_getBalance
${unlockAccountMethod}    personal_unlockAccount
${personalListAccountsMethod}    personal_listAccounts
${getHolderCertMethod}    getHolderCertIDs
${getRequesterCertMethod}    getRequesterCert
${checkRequesterCertMethod}    checkRequesterCert
${addCRLMethod}    addCRL
${queryCRLMethod}    getCRL
# address
${prc720ContractAddr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DREThG43
${certContractAddr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DRv2vsEk
${debugContractAddr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DSfQdUHf
${depositContractAddr}    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM
# comman param
${tokenHolder}    ${null}
${Alice}          ${null}
${Bob}            ${null}
${Carol}          ${null}
${amount}         10000
${fee}            1
${pwd}            1
${duration}       600000000
${gasToken}       WWW
${AliceToken}     ALICE
${BobToken}       BOB
${CarolToken}     CAROL
${AliceTokenID}    ${null}
${BobTokenID}     ${null}
${CarolTokenID}    ${null}
# certificate
${caCertHolder}    ${null}
${powerCertHolder}    ${null}
${userCertHolder}    ${null}
${caCertID}       ${null}
${powerCertID}    ${null}
${userCertID}     ${null}
${caCertBytes}    -----BEGIN CERTIFICATE-----\n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\nB8H3cQht7g48a4RhUw==\n-----END CERTIFICATE-----
${powerCertBytes}    -----BEGIN CERTIFICATE-----\nMIICKTCCAdCgAwIBAgIUOLt0JeImTIplnp5WW3oBoh8nZDIwCgYIKoZIzj0EAwIw\naDELMAkGA1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMRQwEgYDVQQK\nEwtIeXBlcmxlZGdlcjEPMA0GA1UECxMGRmFicmljMRkwFwYDVQQDExBmYWJyaWMt\nY2Etc2VydmVyMB4XDTE5MDgxMzEwMzAwMFoXDTI0MDgxMTEwMzUwMFowWjELMAkG\nA1UEBhMCVVMxFzAVBgNVBAgTDk5vcnRoIENhcm9saW5hMRQwEgYDVQQKEwtIeXBl\ncmxlZGdlcjEPMA0GA1UECxMGY2xpZW50MQswCQYDVQQDEwJsazBZMBMGByqGSM49\nAgEGCCqGSM49AwEHA0IABBlwYvePFCNRfMAfh4L44iVGYl/eyzFWWcSal1CP4mJg\nKxlwuYNWBXewrw8RjXdeArgpCgDJWFt2OhP15+7tO7ujZjBkMA4GA1UdDwEB/wQE\nAwIBBjASBgNVHRMBAf8ECDAGAQH/AgEAMB0GA1UdDgQWBBTzY152TL2rhUD61gVD\nj5GBt2+R0DAfBgNVHSMEGDAWgBQPefDSpvBGlFupTMF4hIQdDeNvgzAKBggqhkjO\nPQQDAgNHADBEAiAjZ36ms0v+/JUnZspOm1ezmX2GHW4LbzHX42e3nUulBgIgCC+D\nNbzNucp3aSAQqJCVAi1iN7Af7ZWobtaZQyDcQY0=\n-----END CERTIFICATE-----\n
