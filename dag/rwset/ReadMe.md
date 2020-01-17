## TxSimulator设计思路

1. 开始创建一个新单元时创建TxSimulator实例
2. 一个单元中包含多笔系统合约的Invoke时，使用同一个TxSimulator实例。接口中的ns为RequestId
3. 单元创建完毕时，需要关闭TxSimulator

## 连续交易涉及问题
除了合约的读写集，还涉及到合约相关的UTXO，如果是连续交易，下一个合约应该可以用上一个合约的写集和生成的UTXO
因为涉及到合约手续费不够或者合约错误，导致写集和UTXO回滚，所以应该提供清除某RequestId的写集和UTXO的方法。
为了防止双花，一个连续交易中，上一个Tx用掉的UTXO，下一个Tx不能再使用，但可以使用上一个Tx生成的新UTXO。
关于读集的版本问题，如果AB是连续交易，B读了A产生的写集，B的读集里面的Version就应该是空？