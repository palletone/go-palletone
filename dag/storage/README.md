Storage封装了对底层数据库对象的增删改查操作，每个操作只针对一个对象，按存储的对象功能不同，分为：
* DagDB
* UTXODB
* StateDB
* IndexDB
* PropertyDb
其中DAGDB存储所有区块（Unit）的信息,当然Unit中包含的Transaction也在其中。
UTXODB是根据Tx中的PaymentMessage而构建的Token UTXO状态的数据库，该数据库中的UTXO是进行软删除。
StateDB是根据Tx中的ContractInvoke中的RWSet而构建的状态数据库，状态数据在数据库中保存时存在版本信息。另外对账户状态，系统配置，合约模板，合约的其他属性等都在该数据库。
IndexDB是因为快速检索的需要而构建的索引数据库。我们可以通过toml中的配置来打开或关闭对某些数据的索引。
PropertyDB是专门为不需要保存到区块链，但是又需要临时存储的数据而设置的数据库。该数据库中包含了MediatorSchedule，GlobalProperty，DynamicProperty等。


## UTXO DB的设计
UTXO DB中，以OutPoint为Key（UTXO_PREFIX），以Utxo为Value存储最基本的UTXO数据。
为了能够方便地址快速访问自己的UTXO，所以需要强制建立索引，索引的Key为：
AddrOutPoint_Prefix+地址.Bytes()+OutPoint.Bytes()
Value 仍然是 OutPoint这个对象的RLP编码

SaveUtxoEntity存储一个UTXO，同时也会存储一个对应Address对应的索引
删除Utxo的时候软删除Utxo，但是会物理删除对应的索引
