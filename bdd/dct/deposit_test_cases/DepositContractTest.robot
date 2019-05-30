*** Settings ***
Resource          publicParams.txt
Library           RequestsLibrary

*** Variables ***
${mediatorAddr_01}    ${EMPTY}
${foundationAddr}    ${EMPTY}
${mediatorAddr_02}    ${EMPTY}
${juryAddr_01}    ${EMPTY}
${developerAddr_01}    ${EMPTY}
${anotherAddr}    ${EMPTY}

*** Test Cases ***
Business_01
    [Documentation]    某节点申请加入mediator-》进入申请列表-》基金会同意-》进入同意列表-》节点加入保证金（足够）-》进入候选列表-》节点申请退出候选列表-》进入退出列表-》基金会同意。此时，所有列表为空
    ${result}    applyBecomeMediator    ${mediatorAddr_01}    #${mediatorAddr_01}节点申请加入列表
    log    ${result}
    ${isIn}    getBecomeMediatorApplyList    ${mediatorAddr_01}    #${mediatorAddr_01}是否在申请列表里面
    log    ${isIn}
    Should Be Equal As Strings    ${isIn}    True    #为true
    ${result}    handleForApplyBecomeMediator    ${foundationAddr}    ${mediatorAddr_01}    #基金会处理列表里的节点（同意）
    log    ${result}
    ${isIn}    getAgreeForBecomeMediatorList    ${mediatorAddr_01}    #${mediatorAddr_01}是否在同意列表里面
    log    ${isIn}
    Should Be Equal As Strings    ${isIn}    True    #为true
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_01}    200000000000    #在同意列表里的节点，可以交付保证金（大于或等于保证金数量）,需要200000000000及以上
    log    ${result}
    ${isIn}    getListForMediatorCandidate    ${mediatorAddr_01}    #${mediatorAddr_01}是否在候选列表里面
    log    ${isIn}
    Should Be Equal As Strings    ${isIn}    True    #为true
    ${result}    getCandidateBalanceWithAddr    ${mediatorAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil    #有余额
    ${result}    applyQuitMediator    ${mediatorAddr_01}    #该节点申请退出mediator候选列表
    log    ${result}
    ${isIn}    getQuitMediatorApplyList    ${mediatorAddr_01}    #获取申请mediator列表里的节点（不为空）
    log    ${isIn}
    Should Be Equal As Strings    ${isIn}    True    #为true
    ${result}    handleForApplyForQuitMediator    ${foundationAddr}    ${mediatorAddr_01}    #基金会处理退出候选列表里的节点（同意）
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${mediatorAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    #    ${result}    balance is nil    #账户地址不存在
    #    getBecomeMediatorApplyList    ${mediatorAddr_01}
    #    ${result}
    #    getAgreeForBecomeMediatorList    ${mediatorAddr_01}
    #    ${result}
    #    getListForMediatorCandidate    ${mediatorAddr_01}
    #    ${result}
    #    getQuitMediatorApplyList    ${mediatorAddr_01}
    #    ${result}

Business_02
    [Documentation]    某节点申请加入mediator-》进入申请列表-》基金会同意-》进入同意列表-》节点加入保证金（足够）-》进入候选列表-》社区节点申请没收改地址所以保证金-》基金会同意，此时，只有同意列表不为空，其他的为空。
    ${result}    applyBecomeMediator    ${mediatorAddr_02}    #节点申请加入列表
    log    ${result}
    @{addressList1}    getBecomeMediatorApplyList    #获取申请加入列表的节点（不为空）
    log    @{addressList1}
    Should Be True    '${mediatorAddr_02}' in @{addressList1}
    ${result}    handleForApplyBecomeMediator    ${foundationAddr}    0    #基金会处理列表里的节点（同意）
    log    ${result}
    @{addressList2}    getAgreeForBecomeMediatorList    #获取同意列表的节点（不为空）
    log    @{addressList2}
    Should Be True    '${mediatorAddr_02}' in @{addressList2}
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_02}    200000000000    #在同意列表里的节点，可以交付保证金（大于或等于保证金数量）,需要200000000000及以上
    log    ${result}
    @{addressList3}    getListForMediatorCandidate    #交付足够保证金后，可加入mediator候选列表（不为空）
    log    @{addressList3}
    Should Be True    '${mediatorAddr_02}' in @{addressList3}
    ${result}    getCandidateBalanceWithAddr    ${mediatorAddr_02}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil    #有余额
    ${result}    applyForForfeitureDeposit    ${anotherAddr}    ${mediatorAddr_02}    200000000000    Mediator    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication    #不为空
    log    ${result}
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${mediatorAddr_02}    #获取该地址保证金账户详情
    log    ${result}
    Should Be Equal    ${result}    balance is nil    #为空
    ${result}    getBecomeMediatorApplyList    #为空
    log    ${result}
    ${result}    getAgreeForBecomeMediatorList    #不为空
    log    ${result}
    ${result}    getListForMediatorCandidate    #为空
    log    ${result}
    ${result}    getListForForfeitureApplication    #为空
    log    ${result}

Business_03
    [Documentation]    某节点想成为jury，直接可交付保证金，但是该节点交的钱不够，所以无法加入jury候选列表，但是可以查询改地址在合约账户的相关信息
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    40000000000    #jury交付保证金，规定至少需要交付1000 0000 0000的，但是现在不够数量，所以无法加入候选列表
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate    #为空
    log    ${resul}

Business_04
    [Documentation]    该节点想成为合约开发者，由于交付保证金足够，所以加入了开发者的候选列表，也可以查询到账户余额信息。
    ${result}    developerPayToDepositContract    ${developerAddr_01}    80000000000    #developer交付保证金，足够数量，80000000000，加入候选列表
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${result}    getListForDeveloperCandidate    #获取developer候选列表里的节点，不为空
    log    ${result}

Business_06
    [Documentation]    该节点是在 Business_03 的基础上，再进行交付一下保证金，但是加起来还是没有达到保证金数量，所以jury候选列表为空，然后该节点提出申请退还一些保证金，基金会同意，退出保证金列表为空，该节点的账户余额会减少。
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    40000000000    #jury交付保证金，规定至少需要交付1000 0000 0000的，但是现在不够数量，所以无法加入候选列表
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate    #为空
    log    ${resul}
    ${result}    juryApplyCashback    ${juryAddr_01}    50000000000    #由于交付保证金时没有足够，所以想退多少就退多少，在余额之内，并没有其他相关操作了
    log    ${result}
    ${result}    getListForCashbackApplication    #不为空
    log    ${result}
    Should Contain    ${result}    ${juryAddr_01}
    ${result}    handleForCashbackApplication    ${foundationAddr}    HandleForJuryApplyCashback    ok    #基金会处理退保证金申请（同意）
    log    ${result}
    ${result}    getListForCashbackApplication    #为空，并且jury余额减少
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil

Business_07
    [Documentation]    该节点在 Business_04 的基础上，申请退出相应保证金，由于退出保证金后，余额不足，会被移除开发者候选列表，账户余额改变，退出列表为空，候选列表为空。
    ${result}    developerApplyCashback    ${developerAddr_01}    50000000000    #developer申请退保证金，由于刚才交了足够保证金数量，加入了候选列表，现在退保证金，退出后，余额不足够，所以需要对其移除出developer候选列表
    log    ${result}
    ${result}    getListForCashbackApplication    #不为空
    log    ${result}
    #    ${result}    ${developerAddr_01}
    ${result}    getListForDeveloperCandidate    #不为空
    log    ${result}
    #    ${result}    ${developerAddr_01}
    ${result}    handleForCashbackApplication    ${foundationAddr}    HandleForDeveloperApplyCashback    ok    #基金会处理退保证金申请（同意）
    log    ${result}
    ${result}    getListForCashbackApplication    #为空，并且developer余额减少
    log    ${result}
    ${result}    getListForDeveloperCandidate    #为空
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil

Business_05
    [Documentation]    在 Business_02 的基础上，被没收节点继续申请加入mediator候选列举，因此该地址只需交付足够保证金即可，但是，由于该节点交付不够，所以无法加入候选类别，并且规定账户也是为空的。
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_02}    100000000000    #在同意列表里的节点，可以交付保证金（大于或等于保证金数量）,需要200000000000及以上
    log    ${result}
    getListForMediatorCandidate    #交付足够保证金后，可加入mediator候选列表（不为空）
    ${result}    getCandidateBalanceWithAddr    ${mediatorAddr_02}    #获取该地址保证金账户详情
    log    ${result}
    Should Be Equal    ${result}    balance is nil    #无余额

Business_08
    [Documentation]    在 Business_02 的基础上，被没收节点继续申请加入mediator候选列举，因此该地址只需交付足够保证金即可
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_02}    250000000000    #在同意列表里的节点，可以交付保证金（大于或等于保证金数量）,需要200000000000及以上
    log    ${result}
    @{addressList3}    getListForMediatorCandidate    #交付足够保证金后，可加入mediator候选列表（不为空）
    log    @{addressList3}
    Should Be True    '${mediatorAddr_02}' in @{addressList3}
    ${result}    getCandidateBalanceWithAddr    ${mediatorAddr_02}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil    #有余额

Business_09
    [Documentation]    在 Business_06 的基础上，该节点继续交付保证金想成为 jury 候选列表的节点，此次交付了足够的保证金，所以成功进入候选列表，并且余额相应的改变。
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    140000000000    #jury交付保证金，规定至少需要交付1000 0000 0000的，但是现在不够数量，所以无法加入候选列表
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate    #不为空
    log    ${resul}
    Should Contain    ${result}    ${juryAddr_01}

Business_10
    [Documentation]    在 Business_09 的基础上，该节点想退还保证金，但是基金会不同意，所以申请列表置空，节点账户不变
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${result}    juryApplyCashback    ${juryAddr_01}    50000000000    #由于交付保证金时没有足够，所以想退多少就退多少，在余额之内，并没有其他相关操作了
    log    ${result}
    ${result}    getListForCashbackApplication    #不为空
    log    ${result}
    Should Contain    ${result}    ${juryAddr_01}
    ${result}    handleForCashbackApplication    ${foundationAddr}    HandleForJuryApplyCashback    no    #基金会处理退保证金申请（同意）
    log    ${result}
    ${result}    getListForCashbackApplication    #为空，但是jury余额不变
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
