*** Settings ***
Resource          publicParams.txt

*** Variables ***
${mediatorAddr_01}    ${EMPTY}
${foundationAddr}    ${EMPTY}
${mediatorAddr_02}    ${EMPTY}
${juryAddr_01}    ${EMPTY}
${developerAddr_01}    ${EMPTY}
${mediatorAddr_03}    ${EMPTY}
${anotherAddr}    ${EMPTY}

*** Test Cases ***
Business_01
    ${result}    applyBecomeMediator    ${mediatorAddr_01}    #mediator 申请加入列表
    log    ${result}
    ${result}    getBecomeMediatorApplyList    #获取申请加入列表的节点（不为空）
    log    ${result}
    ${jsonResult}    to json    ${result}
    Should Contain    ${jsonResult['result']}    ${mediatorAddr_01}
    ${result}    handleForApplyBecomeMediator    ${foundationAddr}    #基金会处理列表里的节点（同意）
    log    ${result}
    ${result}    getAgreeForBecomeMediatorList    #获取同意列表的节点（不为空）
    log    ${result}
    ${jsonResult}    to json    ${result}
    Should Contain    ${jsonResult['result']}    ${mediatorAddr_01}
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_01}    200000000000    #在同意列表里的节点，可以交付保证金（大于或等于保证金数量）,需要200000000000及以上
    log    ${result}
    ${result}    getListForMediatorCandidate    #交付足够保证金后，可加入mediator候选列表（不为空）
    log    ${result}
    ${jsonResult}    to json    ${result}
    Should Contain    ${jsonResult['result']}    ${mediatorAddr_01}
    ${result}    getCandidateBalanceWithAddr    ${mediatorAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    ${jsonResult}    to json    ${result}
    Should Contain    ${jsonResult['result']}    200000000000

Business_02
    ${result}    applyBecomeMediator    ${mediatorAddr_02}    #mediator 申请加入列表
    log    ${result}
    ${result}    getBecomeMediatorApplyList    #获取申请加入列表的节点（不为空）
    log    ${result}
    ${jsonResult}    to json    ${result}
    Should Contain    ${jsonResult['result']}    ${mediatorAddr_02}
    ${result}    handleForApplyBecomeMediator    ${foundationAddr}    #基金会处理列表里的节点（同意）
    log    ${result}
    ${result}    getAgreeForBecomeMediatorList    #获取同意列表的节点（不为空）
    log    ${result}
    ${jsonResult}    to json    ${result}
    Should Contain    ${jsonResult['result']}    ${mediatorAddr_02}
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_02}    200000000000    #在同意列表里的节点，可以交付保证金（大于或等于保证金数量）,需要200000000000及以上
    log    ${result}
    ${result}    getListForMediatorCandidate    #交付足够保证金后，可加入mediator候选列表（不为空）
    log    ${result}
    ${jsonResult}    to json    ${result}
    Should Contain    ${jsonResult['result']}    ${mediatorAddr_02}
    ${result}    getCandidateBalanceWithAddr    ${mediatorAddr_02}    #获取该地址保证金账户详情
    log    ${result}
    ${result}    applyQuitMediator    ${mediatorAddr_02}    #申请退出mediator候选列表
    log    ${result}
    ${result}    getQuitMediatorApplyList    #获取申请mediator列表里的节点（不为空）
    log    ${result}
    ${jsonResult}    to json    ${result}
    Should Contain    ${jsonResult['result']}    ${mediatorAddr_02}
    ${result}    handleForApplyForQuitMediator    ${foundationAddr}    #基金会处理退出候选列表里的节点（同意）
    log    ${result}
    ${result}    getQuitMediatorApplyList    #为空
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${mediatorAddr_02}    #获取该地址保证金账户详情
    log    ${result}

Business_03
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    90000000000    #jury交付保证金，规定至少需要交付1000 0000 0000的，但是现在不够数量，所以无法加入候选列表
    log    ${resul}
    ${resul}    getListForJuryCandidate    #为空
    log    ${resul}

Business_04
    ${result}    developerPayToDepositContract    ${developerAddr_01}    80000000000    #developer交付保证金，足够数量，80000000000，加入候选列表
    log    ${result}
    ${result}    getListForDeveloperCandidate    #获取developer候选列表里的节点，不为空
    log    ${result}

Business_05
    ${result}    mediatorApplyCashback    ${mediatorAddr_01}    50000000000    #mediator 申请保证金，但是由于对mediator的特殊处理：即要么退出部分保证金，剩余保证金依然足够；要么全部退出。显然当前退出保证金不行
    log    ${result}
    ${result}    getListForCashbackApplication    #为空
    log    ${result}

Business_06
    ${result}    juryApplyCashback    ${juryAddr_01}    500000000000    #由于交付保证金时没有足够，所以想退多少就退多少，在余额之内，并没有其他相关操作了
    log    ${result}
    ${result}    getListForCashbackApplication    #不为空
    log    ${result}
    ${result}    handleForCashbackApplication    ${foundationAddr}    HandleForJuryApplyCashback    #基金会处理退保证金申请（同意）
    log    ${result}
    ${result}    getListForCashbackApplication    #为空，并且jury余额减少
    log    ${result}

Business_07
    ${result}    developerApplyCashback    ${developerAddr_01}    50000000000    #developer申请退保证金，由于刚才交了足够保证金数量，加入了候选列表，现在退保证金，退出后，余额不足够，所以需要对其移除出developer候选列表
    log    ${result}
    ${result}    getListForCashbackApplication    #不为空
    log    ${result}
    \    getListForDeveloperCandidate    #不为空
    ${result}    handleForCashbackApplication    ${foundationAddr}    HandleForDeveloperApplyCashback    #基金会处理退保证金申请（同意）
    log    ${result}
    ${result}    getListForCashbackApplication    #为空，并且developer余额减少
    log    ${result}
    ${result}    getListForDeveloperCandidate    #为空
    log    ${result}

Business_08
    ${result}    applyBecomeMediator    ${mediatorAddr_03}    #mediator 申请加入列表
    log    ${result}
    ${result}    getBecomeMediatorApplyList    #获取申请加入列表的节点（不为空）
    log    ${result}
    ${result}    handleForApplyBecomeMediator    ${foundationAddr}    #基金会处理列表里的节点（同意）
    log    ${result}
    ${result}    getAgreeForBecomeMediatorList    #获取同意列表的节点（不为空）
    log    ${result}
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_03}    300000000000    #在同意列表里的节点，可以交付保证金（大于或等于保证金数量）,需要200000000000及以上
    log    ${result}
    ${result}    getListForMediatorCandidate    #交付足够保证金后，可加入mediator候选列表（不为空）
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${mediatorAddr_03}    #获取该地址保证金账户详情
    log    ${result}
    ${result}    applyForForfeitureDeposit    ${anotherAddr}    ${mediatorAddr_03}    300000000000    Mediator    #某个地址申请没收该mediator保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication    #不为空
    log    ${result}
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getListForForfeitureApplication    #为空
    log    ${result}
    ${result}    getListForMediatorCandidate    #为空
    log    ${result}
