# Token Engine介绍

TokenEngine来自于btcd的txscript模块，该模块实现了比特币的堆栈引擎。对于txscript项目做了如下修改：
1. 屏蔽了隔离见证相关的代码。
2. 修改了Tx结构，采用了Tx->Message->Input/Output这样的三层结构
3. 增加了OPCODE：OP_JURY_REDEEM_EQUAL   = 0xc8 // 200  用于从状态数据库获得一个合约的陪审团赎回代码并判断是否相等。
4. 增加了SigHashRaw          SigHashType = 0x4 //直接对构造好的不包含任何签名信息的Tx签名
5. 增加了付款到合约和从合约付款出去的支持。


