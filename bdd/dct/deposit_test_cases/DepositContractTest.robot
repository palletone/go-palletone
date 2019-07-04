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
    [Documentation]    mediator 交付 50 0000 0000 0000 及以上才可以加入候选列表
    ...
    ...    某节点申请加入mediator-》进入申请列表-》基金会同意-》进入同意列表-》节点加入保证金（足够）-》进入候选列表-》节点增加保证金-》节点申请退出部分保证金-》基金会同意-》节点申请退出候选列表-》进入退出列表-》基金会同意。
    ${amount}    getBalance    ${mediatorAddr_01}    PTN
    log    ${amount}    #999949995
    #    Should Be Equal As Numbers
    Should Be Equal As Numbers    ${amount}    999949995
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    0
    ${result}    applyBecomeMediator    ${mediatorAddr_01}    #节点申请加入列表    #1
    log    ${result}
    ${addressMap1}    getBecomeMediatorApplyList
    log    ${addressMap1}
    Dictionary Should Contain Key    ${addressMap1}    ${mediatorAddr_01}    #有该节点
    ${result}    handleForApplyBecomeMediator    ${foundationAddr}    ${mediatorAddr_01}    Ok    #基金会处理列表里的节点（同意）    #1
    log    ${result}
    ${addressMap2}    getAgreeForBecomeMediatorList
    log    ${addressMap2}
    Dictionary Should Contain Key    ${addressMap2}    ${mediatorAddr_01}    #有该节点
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_01}    3000000000    #在同意列表里的节点，可以交付保证金    #这里交的数量不是规定的保证金数量，导致无法加入候选列表，并且相应保证金退还该地址    #1
    log    ${result}
    ${amount}    getBalance    ${mediatorAddr_01}    PTN    #499999992
    log    ${amount}    #999949992
    Should Be Equal As Numbers    ${amount}    999949992
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    0
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Not Contain Key    ${addressMap3}    ${mediatorAddr_01}    #无该节点
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_01}    ${medDepositAmount}    #在同意列表里的节点，可以交付保证金    #51
    log    ${result}
    ${amount}    getBalance    ${mediatorAddr_01}    PTN
    log    ${amount}    #999949941
    Should Be Equal As Numbers    ${amount}    999949941
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    50
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Contain Key    ${addressMap3}    ${mediatorAddr_01}    #有该节点
    ${resul}    getListForJuryCandidate    #mediator自动称为jury
    Dictionary Should Contain Key    ${resul}    ${mediatorAddr_01}    #有该节点
    log    ${resul}
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_01}    #获取该地址保证金账户详情
    log    ${mDeposit}
    Should Not Be Equal    ${mDeposit["balance"]}    ${0}    #有余额
    ${result}    applyQuitMediator    ${mediatorAddr_01}    MediatorApplyQuit    #该节点申请退出mediator候选列表    #1
    log    ${result}
    ${addressMap4}    getQuitMediatorApplyList
    log    ${addressMap4}
    Dictionary Should Contain Key    ${addressMap4}    ${mediatorAddr_01}    #有该节点
    ${result}    handleForApplyForQuitMediator    ${foundationAddr}    ${mediatorAddr_01}    Ok    HandleForApplyQuitMediator    #基金会处理退出候选列表里的节点（同意）
    ...    #1
    log    ${result}
    ${amount}    getBalance    ${mediatorAddr_01}    PTN    #99,999,970‬
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    999949989
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    0
    ${resul}    getListForJuryCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Not Contain Key    ${resul}    ${mediatorAddr_01}    #无该节点
    log    ${resul}
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_01}    #获取该地址保证金账户详情
    log    ${mDeposit}
    Should Be Equal    ${mDeposit["balance"]}    ${0}    #账户地址存在
    ${result}    getBecomeMediatorApplyList
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_01}    #无该节点
    ${result}    getAgreeForBecomeMediatorList
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${mediatorAddr_01}    #有该节点
    ${result}    getListForMediatorCandidate
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_01}    #无该节点
    ${result}    getQuitMediatorApplyList
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${mediatorAddr_01}    #无该节点

Business_02
    [Documentation]    没收mediator节点
    ${result}    applyBecomeMediator    ${mediatorAddr_02}    #节点申请加入列表
    log    ${result}
    ${addressMap1}    getBecomeMediatorApplyList
    log    ${addressMap1}
    Dictionary Should Contain Key    ${addressMap1}    ${mediatorAddr_02}    #有该节点
    ${result}    handleForApplyBecomeMediator    ${foundationAddr}    ${mediatorAddr_02}    Ok    #基金会处理列表里的节点（同意）
    log    ${result}
    ${addressMap2}    getAgreeForBecomeMediatorList
    log    ${addressMap2}
    Dictionary Should Contain Key    ${addressMap2}    ${mediatorAddr_02}    #有该节点
    ${result}    mediatorPayToDepositContract    ${mediatorAddr_02}    ${medDepositAmount}    #在同意列表里的节点，可以交付保证金（大于或等于保证金数量）,需要200000000000及以上
    log    ${result}
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Contain Key    ${addressMap3}    ${mediatorAddr_02}    #有该节点
    ${resul}    getListForJuryCandidate    #mediator自动称为jury
    Dictionary Should Contain Key    ${resul}    ${mediatorAddr_02}    #有该节点
    log    ${resul}
    ${mDeposit}    getMediatorDepositWithAddr    ${mediatorAddr_02}    #获取该地址保证金账户详情
    log    ${mDeposit}
    Should Not Be Equal    ${mDeposit["balance"]}    ${0}    #有余额
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${mediatorAddr_02}    Mediator    nothing to do    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${mediatorAddr_02}    #有该节点
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${mediatorAddr_02}    Ok    #基金会处理（同意），这是会移除mediator出候选列表
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

Business_03
    [Documentation]    jury 交付 1000000 0000 0000 及以上才可以加入候选列表
    ${resul}    juryPayToDepositContract    ${juryAddr_01}    1000000000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_01}    #获取该地址保证金账户详情
    log    ${result}    #余额为100000000000000
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_01}    #候选列表有该地址
    log    ${resul}
    ${result}    applyQuitMediator    ${juryAddr_01}    JuryApplyQuit    #该节点申请退出mediator候选列表
    log    ${result}
    ${addressMap4}    getQuitMediatorApplyList    #获取申请mediator列表里的节点（不为空）
    log    ${addressMap4}
    Dictionary Should Contain Key    ${addressMap4}    ${juryAddr_01}
    ${result}    handleForApplyForQuitMediator    ${foundationAddr}    ${juryAddr_01}    Ok    HandleForApplyQuitJury    #基金会处理退出候选列表里的节点（同意）
    log    ${result}
    ${resul}    getListForJuryCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Not Contain Key    ${resul}    ${juryAddr_01}
    log    ${resul}
    ${result}    getQuitMediatorApplyList    #为空
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${juryAddr_01}

Business_04
    [Documentation]    没收jury节点
    ${resul}    juryPayToDepositContract    ${juryAddr_02}    1000000000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_02}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForJuryCandidate
    Dictionary Should Contain Key    ${resul}    ${juryAddr_02}    #候选列表有该地址
    log    ${resul}
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${juryAddr_02}    Jury    nothing to do    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${juryAddr_02}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${juryAddr_02}    Ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${juryAddr_02}
    log    ${result}
    Should Be Equal    ${result}    balance is nil    #不为空
    ${resul}    getListForJuryCandidate
    Dictionary Should Not Contain Key    ${resul}    ${juryAddr_02}    #候选列表无该地址
    log    ${resul}

Business_05
    [Documentation]    dev 交付 10000 0000 0000 及以上才可以加入候选列表
    ${resul}    developerPayToDepositContract    ${developerAddr_01}    100000000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_01}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_01}    #候选列表无该地址
    log    ${resul}
    ${result}    applyQuitMediator    ${developerAddr_01}    DeveloperApplyQuit    #该节点申请退出mediator候选列表
    log    ${result}
    ${addressMap4}    getQuitMediatorApplyList    #获取申请mediator列表里的节点（不为空）
    log    ${addressMap4}
    Dictionary Should Contain Key    ${addressMap4}    ${developerAddr_01}
    ${result}    handleForApplyForQuitMediator    ${foundationAddr}    ${developerAddr_01}    Ok    HandleForApplyQuitDev    #基金会处理退出候选列表里的节点（同意）
    log    ${result}
    ${resul}    getListForDeveloperCandidate    #mediator退出候选列表，则移除该jury
    Dictionary Should Not Contain Key    ${resul}    ${developerAddr_01}
    log    ${resul}
    ${result}    getQuitMediatorApplyList    #为空
    log    ${result}
    Dictionary Should Not Contain Key    ${result}    ${developerAddr_01}

Business_06
    [Documentation]    没收dev节点
    ${resul}    developerPayToDepositContract    ${developerAddr_02}    100000000
    log    ${resul}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_02}    #获取该地址保证金账户详情
    log    ${result}
    Should Not Be Equal    ${result}    balance is nil
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Contain Key    ${resul}    ${developerAddr_02}    #候选列表无该地址
    log    ${resul}
    ${result}    applyForForfeitureDeposit    ${foundationAddr}    ${developerAddr_02}    Developer    nothing to do    #某个地址申请没收该节点保证金（全部）
    log    ${result}
    ${result}    getListForForfeitureApplication
    log    ${result}
    Dictionary Should Contain Key    ${result}    ${developerAddr_02}    #没收列表有该地址
    ${result}    handleForForfeitureApplication    ${foundationAddr}    ${developerAddr_02}    Ok    #基金会处理（同意），这是会移除mediator出候选列表
    log    ${result}
    ${result}    getCandidateBalanceWithAddr    ${developerAddr_02}
    log    ${result}
    Should Be Equal    ${result}    balance is nil    #不为空
    ${resul}    getListForDeveloperCandidate
    Dictionary Should Not Contain Key    ${resul}    ${developerAddr_02}    #候选列表无该地址
    log    ${resul}

Business_07
    [Documentation]    创建新Token，使用新Token交付保证金，由于保证金只支持PTN，所以交保证金失败
    ...
    ...    DPT+10E7XBWEBT0K2H2GE0O
    ...
    ...    DPT+10102JC6CQU8OK204BXA
    ${amount}    getBalance    ${foundationAddr}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    999950041
    ${amount}    getBalance    ${mediatorAddr_02}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9948
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    0
    ${result}    createToken    ${mediatorAddr_02}    #1
    ${assetId}    ccquery
    log    ${assetId}
    ${amount}    getBalance    ${mediatorAddr_02}    ${assetId}
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    1000000000
    ${result}    invokeToken    ${mediatorAddr_02}    ${assetId}    #在同意列表里的节点，可以交付保证金    #这里交的数量不是规定的保证金数量，导致无法加入候选列表，并且相应保证金退还该地址    #1
    log    ${result}
    ${addressMap3}    getListForMediatorCandidate
    log    ${addressMap3}
    Dictionary Should Not Contain Key    ${addressMap3}    ${mediatorAddr_02}    #无该节点
    ${amount}    getBalance    ${mediatorAddr_02}    ${assetId}
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    1000000000
    ${amount}    getBalance    ${mediatorAddr_02}    PTN
    log    ${amount}
    Should Be Equal As Numbers    ${amount}    9946
    ${amount}    getBalance    PCGTta3M4t3yXu8uRgkKvaWd2d8DR32W9vM    PTN
    log    ${amount}    #0
    Should Be Equal As Numbers    ${amount}    0
