*** Settings ***
Library           Collections
Library           BuiltIn

*** Test Cases ***
test
    ${dict}=    Create Dictionary    a    1    b    2
    ${status}    ${value}=    Run Keyword And Ignore Error    Get From Dictionary    ${dict}    result
    Run Keyword If    '${status}' == 'PASS'    Log    "111"
    ${hosts}=    Create List
    Append To List    ${hosts}    http://localhost:80
    Should Not Be Empty    ${hosts}

test1
    ${res}=    Set Variable    {\"item\":\"transaction_info\",\"info\":{\"tx_hash\":\"0xde2ac615114273107dc57538616af0ec09d4e9f9386c532d93cf4f1dd7d562df\",\"tx_size\":556,\"payment\":[{\"inputs\":[{\"txid\":\"0x327027b1872df05893a2259f391cdf78895cf9d7149ce232f031fc961ee6ecf3\",\"message_index\":0,\"out_index\":0,\"unlock_script\":\"3045022100f82f8933df919e99a78f5cf6836fa7e36186ba1c6445469af7d04728befc133e02202f9e16c73ba8fbd694d9110b8d0af48f4e43ffd6ec35d1781d7f2bb36ae3f5b301 020e132c8e4591d14f0ecdcfc78af0c775ea81956179087c195fcdc1b8c218288a\",\"from_address\":\"\"}],\"outputs\":[{\"amount\":99999999999999888,\"asset\":\"PTN\",\"to_address\":\"P1AUy2NU3VFTc1TcH8LowVbv8awbec2Z3E9\",\"lock_script\":\"OP_DUP OP_HASH160 680321126afed42f696138339461ed91863b9835 OP_EQUALVERIFY OP_CHECKSIG\"}],\"locktime\":0,\"number\":0}],\"fee\":0,\"account_state_update\":null,\"data\":[],\"contract_tpl\":null,\"contract_deploy\":null,\"contract_invoke\":{\"row_number\":2,\"contract_id\":\"PCexAFX1Ldvev7d4dkxBhwf1AvKwa3jiW99\",\"args\":[\"{\\\"invoke_address\\\":\\\"P1AUy2NU3VFTc1TcH8LowVbv8awbec2Z3E9\\\",\\\"invoke_tokens\\\":null,\\\"invoke_fees\\\":{\\\"amount\\\":1,\\\"asset\\\":\\\"PTN\\\"}}\",\"\",\"testPutGlobalState\",\"state2\",\"state2\"],\"read_set\":\"[]\",\"write_set\":\"[]\",\"payload\":\"\",\"error_code\":0,\"error_message\":\"Chaincode Error:Only system contract can call this function.\"},\"contract_stop\":null,\"signature\":null,\"install_request\":null,\"deploy_request\":null,\"invoke_request\":{\"row_number\":1,\"contract_addr\":\"PCexAFX1Ldvev7d4dkxBhwf1AvKwa3jiW99\",\"Args\":[\"testPutGlobalState\",\"state2\",\"state2\"],\"timeout\":0},\"stop_request\":null,\"unit_hash\":\"0x7a9041620187521b7011e1bc4d10a3e136112367ba65b1574904eca2be5c85ae\",\"unit_height\":33,\"timestamp\":\"2019-07-16T14:45:18+08:00\",\"tx_index\":1},\"hex\":\"\"}
    ${errCode}=    Evaluate    re.findall('\"error_code\":(\\d*)', '${res}')    re
    ${errMsg}=    Evaluate    re.findall('\"error_message\":\"([^"]*)\"', '${res}')    re
    ${len}=    Get Length    ${errCode}
    ${errCode}=    Run Keyword If    ${len}>=1    Get From List    ${errCode}    0
    ${len}=    Get Length    ${errMsg}
    ${errMsg}=    Run Keyword If    ${len}>=1    Get From List    ${errMsg}    0
    Run Keyword If    ${errCode}!=0    Fail    ${errMsg}
