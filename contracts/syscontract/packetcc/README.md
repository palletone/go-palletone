## 红包合约
### 红包的情形
#### 1.随机红包
MinAmount<MaxAmount
#### 2.固定金额红包
MinAmount==MaxAmount
#### 3.领取时指定金额红包
红包定时中isConstant=true
只有这种红包允许多Token,每种Token的余额是记录在Packet.Tokens对象中