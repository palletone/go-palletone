*** Settings ***
Resource          publicParams.txt
Library           RequestsLibrary

*** Variables ***
${mediatorAddr_01}    ${EMPTY}
${foundationAddr}    ${EMPTY}
${mediatorAddr_02}    ${EMPTY}
${juryAddr_01}    ${EMPTY}
${developerAddr_01}    ${EMPTY}
${juryAddr_02}    ${EMPTY}
${developerAddr_02}    ${EMPTY}

*** Test Cases ***
Business_01
    [Documentation]    mediator 交付 2000 0000 0000 及以上才可以加入候选列表
    ...
    ...    某节点申请加入mediator-》进入申请列表-》基金会同意-》进入同意列表-》节点加入保证金（足够）-》进入候选列表-》节点增加保证金-》节点申请退出部分保证金-》基金会同意-》节点申请退出候选列表-》进入退出列表-》基金会同意。
    ${result}    applyBecomeMediator    ${mediatorAddr_01}    #节点申请加入列表
    log    ${result}
    ${addressMap1}    getBecomeMediatorApplyList    #获取申请加入列表的节点（不为空）
    log    ${addressMap1}
    Dictionary Should Contain Key    ${addressMap1}    ${mediatorAddr_01}
    ${result}    handleForApplyBecomeMediator    ${foundationAddr}    ${mediatorAddr_01}    ok    #基金会处理列表里的节点（同意）
    log    ${result}
    ${addressMap2}    getAgreeForBecomeMediatorList    #获取同意列表的节点（不为空）
    log    ${addressMap2}
    Dictionary Should Contain Key    ${addressMap2}    ${mediatorAddr_01}
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_01}    ${medDepositAmount}    #在同意列表里的节点，可以交付保证金（大于或等于保证金数量）,需要200000000000及以上
    log    ${result}
    ${addressMap3}    getListForMediatorCandidate    #交付足够保证金后，可加入mediator候选列表（不为空）
    log    ${addressMap3}
    Dictionary Should Contain Key    ${addressMap3}    ${mediatorAddr_01}
    ${resul}    getListForJuryCandidate    #mediator自动称为jury
    Dictionary Should Contain Key    ${resul}    ${mediatorAddr_01}
    log    ${resul}
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_01}    #获取该地址保证金账户详情
    log    ${mDeposit}
    Should Not Be Equal    ${mDeposit["balance"]}    ${0}    #有余额
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_01}    ${medDepositAmount}    #增加保证金
    log    ${result}
    ${result}    mediatorApplyCashback    ${mediatorAddr_01}    ${medDepositAmount}    #申请退出部分保证金
    log    ${result}
    ${addressMap5}    getListForCashbackApplication
    log    ${addressMap5}
    Dictionary Should Contain Key    ${addressMap5}    ${mediatorAddr_01}    #mediatorAddr_01
    ${result}    handleForMediatorApplyCashback    ${foundationAddr}    ${mediatorAddr_01}    ok    #基金会处理退保证金列表里的节点（同意）
    log    ${result}
    ${result}    applyQuitMediator    ${mediatorAddr_01}    #该节点申请退出mediator候选列表
    log    ${result}
    ${addressMap4}    getQuitMediatorApplyList    #获取申请mediator列表里的节点（不为空）
    log    ${addressMap4}
    Dictionary Should Contain Key    ${addressMap4}    ${mediatorAddr_01}
    ${result}    handleForApplyForQuitMediator    ${foundationAddr}    ${mediatorAddr_01}    ok    #基金会处理退出候选列表里的节点（同意）
    log    ${result}
    ${resul}    getListForJuryCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Not Contain Key    ${resul}    ${mediatorAddr_01}
    log    ${resul}
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_01}    #获取该地址保证金账户详情
    log    ${mDeposit}
    Should Be Equal    ${mDeposit["balance"]}    ${0}    #账户地址不存在
    ${result}    getBecomeMediatorApplyList    #为空
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_01}
    ${result}    getAgreeForBecomeMediatorList    #为空
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${mediatorAddr_01}
    ${result}    getListForMediatorCandidate    #为空
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_01}
    ${result}    getQuitMediatorApplyList    #为空
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_01}

Business_02
    [Documentation]    没收mediator节点
    ${result}    applyBecomeMediator    ${mediatorAddr_02}    #节点加入申请列表
    log    ${result}
    ${addressMap1}    getBecomeMediatorApplyList
    log    ${addressMap1}
    Dictionary Should Contain Key    ${addressMap1}    ${mediatorAddr_02}    #申请列表有该地址
    ${result}    handleForApplyBecomeMediator    ${foundationAddr}    ${mediatorAddr_02}    ok    #基金会处理列表里的节点（同意）
    log    ${result}
    ${result}    getBecomeMediatorApplyList
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_02}    #申请列表无该地址
    ${addressMap2}    getAgreeForBecomeMediatorList
    log    ${addressMap2}
    Dictionary Should Contain Key    ${addressMap2}    ${mediatorAddr_02}    #同意列表有该地址
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_02}    210000000000    #节点交付2100 0000 0000
    log    ${result}
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Contain Key    ${addressMap3}    ${mediatorAddr_02}    #候选列表有该地址
    ${resul}    getListForJuryCandidate    #mediator自动称为jury
    Dictionary Should Contain Key    ${resul}    ${mediatorAddr_02}    #jury候选列表有该地址
    log    ${resul}
    ${result}    getMediatorDepositWithAddr    ${mediatorAddr_02}    #获取该地址保证金账户详情
    log    ${result}    #余额为 2100 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${mediatorAddr_02}    100000000000    Mediator    nothing to do
    ...    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${mediatorAddr_02}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${mediatorAddr_02}    ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getMediatorDepositWithAddr    ${mediatorAddr_02}
    log    ${result}    #余额为 0
    Should Not Be Equal    ${result}    balance is nil    #不为空
    ${result}    getAgreeForBecomeMediatorList
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${mediatorAddr_02}    #同意列表有该地址
    ${result}    getListForMediatorCandidate
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_02}    #候选列表无该地址
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_02}    #没收列表无该地址
    ${resul}    getListForJuryCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Not Contain Key    ${resul}    ${mediatorAddr_02}    #jury候选列表无该地址
    log    ${resul}
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_02}    350000000000    #节点交付3500 0000 0000
    log    ${result}
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Contain Key    ${addressMap3}    ${mediatorAddr_02}    #候选列表有该地址
    ${resul}    getListForJuryCandidate    #mediator自动称为jury
    Dictionary Should Contain Key    ${resul}    ${mediatorAddr_02}    #jury候选列表有该地址
    log    ${resul}
    ${result}    getMediatorDepositWithAddr    ${mediatorAddr_02}    #获取该地址保证金账户详情
    log    ${result}    #余额为 3500 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${mediatorAddr_02}    100000000000    Mediator    nothing to do
    ...    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${mediatorAddr_02}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${mediatorAddr_02}    ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getMediatorDepositWithAddr    ${mediatorAddr_02}
    log    ${result}    #余额为 2500 0000 0000
    Should Not Be Equal    ${result}    balance is nil    #不为空
    ${result}    getAgreeForBecomeMediatorList
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${mediatorAddr_02}    #同意列表有该地址
    ${result}    getListForMediatorCandidate
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${mediatorAddr_02}    #候选列表无该地址
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_02}    #没收列表无该地址
    ${resul}    getListForJuryCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Contain Key    ${resul}    ${mediatorAddr_02}    #jury候选列表无该地址
    log    ${resul}

Business_03
    [Documentation]    jury 交付 1000 0000 0000 及以上才可以加入候选列表
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    40000000000    #jury交付400 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为400 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Not Contain Key    ${resul}    ${juryAddr_01}    #候选列表无该地址
    log    ${resul}
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    40000000000    #jury交付400 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为800 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Not Contain Key    ${resul}    ${juryAddr_01}    #候选列表无该地址
    log    ${resul}
    ${result}    juryApplyCashback    ${juryAddr_01}    50000000000    #jury提取500 0000 0000
    log    ${result}
    ${result}    getListForCashbackApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${juryAddr_01}    #申请列表有该地址
    ${result}    handleForCashbackApplication    ${foundationAddr}    HandleForJuryApplyCashback    ${juryAddr_01}    ok    #基金会处理退保证金申请（同意）
    log    ${result}
    ${result}    getListForCashbackApplication
    Dictionary Should Not Contain Key    ${result}    ${juryAddr_01}    #申请列表无该地址
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为300 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    80000000000    #jury交付800 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为1100 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_01}    #候选列表有该地址
    log    ${resul}
    ${result}    juryApplyCashback    ${juryAddr_01}    50000000000    #jury提取500 0000 0000
    log    ${result}
    ${result}    getListForCashbackApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${juryAddr_01}    #申请列表有该地址
    ${result}    handleForCashbackApplication    ${foundationAddr}    HandleForJuryApplyCashback    ${juryAddr_01}    ok    #基金会处理退保证金申请（同意）
    log    ${result}
    ${result}    getListForCashbackApplication
    Dictionary Should Not Contain Key    ${result}    ${juryAddr_01}    #申请列表无该地址
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为600 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Not Contain Key    ${resul}    ${juryAddr_01}    #候选列表无该地址
    log    ${resul}
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    80000000000    #jury交付800 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为1400 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_01}    #候选列表有该地址
    log    ${resul}
    ${result}    juryApplyCashback    ${juryAddr_01}    40000000000    #jury提取400 0000 0000
    log    ${result}
    ${result}    getListForCashbackApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${juryAddr_01}    #申请列表有该地址
    ${result}    handleForCashbackApplication    ${foundationAddr}    HandleForJuryApplyCashback    ${juryAddr_01}    ok    #基金会处理退保证金申请（同意）
    log    ${result}
    ${result}    getListForCashbackApplication
    Dictionary Should Not Contain Key    ${result}    ${juryAddr_01}    #申请列表无该地址
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为1000 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_01}    #候选列表有该地址
    log    ${resul}

Business_04
    [Documentation]    没收jury节点
    ${resul}    juryPayToDepositContract    ${juryAddr_02}    140000000000    #jury交付1400 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_02}    #获取该地址保证金账户详情
    log    ${result}    #余额为1400 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_02}    #候选列表有该地址
    log    ${resul}
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${juryAddr_02}    80000000000    Jury    nothing to do
    ...    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${juryAddr_02}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${juryAddr_02}    ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_02}
    log    ${result}    #余额为 600 0000 0000
    Should Not Be Equal    ${result}    balance is nil    #不为空
    ${resul}    getListForJuryCandidate
    Dictionary Should Not Contain Key    ${resul}    ${juryAddr_02}    #候选列表无该地址
    log    ${resul}
    ${resul}    juryPayToDepositContract    ${juryAddr_02}    140000000000    #jury交付1400 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_02}    #获取该地址保证金账户详情
    log    ${result}    #余额为2000 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_02}    #候选列表有该地址
    log    ${resul}
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${juryAddr_02}    80000000000    Jury    nothing to do
    ...    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${juryAddr_02}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${juryAddr_02}    ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_02}
    log    ${result}    #余额为 1200 0000 0000
    Should Not Be Equal    ${result}    balance is nil    #不为空
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_02}    #候选列表无该地址
    log    ${resul}

Business_05
    [Documentation]    dev 交付 800 0000 0000 及以上才可以加入候选列表
    ${resul}    developerPayToDepositContract    ${developerAddr_01}    40000000000    #dev交付400 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为400 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Not Contain Key    ${resul}    ${developerAddr_01}    #候选列表无该地址
    log    ${resul}
    ${resul}    developerPayToDepositContract    ${developerAddr_01}    30000000000    #dev交付300 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为700 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Not Contain Key    ${resul}    ${developerAddr_01}    #候选列表无该地址
    log    ${resul}
    ${result}    developerApplyCashback    ${developerAddr_01}    50000000000    #dev提取500 0000 0000
    log    ${result}
    ${result}    getListForCashbackApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${developerAddr_01}    #申请列表有该地址
    ${result}    handleForCashbackApplication    ${foundationAddr}    HandleForDeveloperApplyCashback    ${developerAddr_01}    ok    #基金会处理退保证金申请（同意）
    log    ${result}
    ${result}    getListForCashbackApplication
    Dictionary Should Not Contain Key    ${result}    ${developerAddr_01}    #申请列表无该地址
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为200 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    developerPayToDepositContract    ${developerAddr_01}    80000000000    #dev交付800 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为1000 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_01}    #候选列表有该地址
    log    ${resul}
    ${result}    developerApplyCashback    ${developerAddr_01}    50000000000    #dev提取500 0000 0000
    log    ${result}
    ${result}    getListForCashbackApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${developerAddr_01}    #申请列表有该地址
    ${result}    handleForCashbackApplication    ${foundationAddr}    HandleForDeveloperApplyCashback    ${developerAddr_01}    ok    #基金会处理退保证金申请（同意）
    log    ${result}
    ${result}    getListForCashbackApplication
    Dictionary Should Not Contain Key    ${result}    ${developerAddr_01}    #申请列表无该地址
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为500 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Not Contain Key    ${resul}    ${developerAddr_01}    #候选列表无该地址
    log    ${resul}
    ${resul}    developerPayToDepositContract    ${developerAddr_01}    80000000000    #dev交付800 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为1300 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_01}    #候选列表有该地址
    log    ${resul}
    ${result}    developerApplyCashback    ${developerAddr_01}    40000000000    #dev提取400 0000 0000
    log    ${result}
    ${result}    getListForCashbackApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${developerAddr_01}    #申请列表有该地址
    ${result}    handleForCashbackApplication    ${foundationAddr}    HandleForDeveloperApplyCashback    ${developerAddr_01}    ok    #基金会处理退保证金申请（同意）
    log    ${result}
    ${result}    getListForCashbackApplication
    Dictionary Should Not Contain Key    ${result}    ${developerAddr_01}    #申请列表无该地址
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为900 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_01}    #候选列表有该地址
    log    ${resul}

Business_06
    [Documentation]    没收dev节点
    ${resul}    developerPayToDepositContract    ${developerAddr_02}    140000000000    #dev交付1400 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_02}    #获取该地址保证金账户详情
    log    ${result}    #余额为1400 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_02}    #候选列表有该地址
    log    ${resul}
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${developerAddr_02}    80000000000    Developer    nothing to do
    ...    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${developerAddr_02}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${developerAddr_02}    ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_02}
    log    ${result}    #余额为 600 0000 0000
    Should Not Be Equal    ${result}    balance is nil    #不为空
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Not Contain Key    ${resul}    ${developerAddr_02}    #候选列表无该地址
    log    ${resul}
    ${resul}    developerPayToDepositContract    ${developerAddr_02}    140000000000    #dev交付1400 0000 0000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_02}    #获取该地址保证金账户详情
    log    ${result}    #余额为2000 0000 0000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_02}    #候选列表有该地址
    log    ${resul}
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${developerAddr_02}    80000000000    Developer    nothing to do
    ...    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${developerAddr_02}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${developerAddr_02}    ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_02}
    log    ${result}    #余额为 1200 0000 0000
    Should Not Be Equal    ${result}    balance is nil    #不为空
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_02}    #候选列表无该地址
    log    ${resul}
