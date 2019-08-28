## 准备

1.安装go

## 下载源码、编译

### go-ethereum

1.下载 go-ethereum 代码

```
go get -u github.com/ethereum/go-ethereum
```

2.编译出 geth 可执行程序

```
cd F:\work\src\github.com\ethereum\go-ethereum\cmd\geth
go build
```


## 启动

+ 启动 geth

```
cd F:\work\src\github.com\ethereum\go-ethereum\cmd\geth
.\geth.exe --datadir "d:\gethtest" --testnet console
```

--datadir 是区块数据目录， --testnet 是指定测试链, console 是控制台环境
去掉参数 --testnet 即是正式链


+ 启动另一个控制台环境

先启动了 geth 才能执行

```
cd F:\work\src\github.com\ethereum\go-ethereum\cmd\geth
.\geth.exe attach \\.\pipe\geth.ipc 
```
\\.\pipe\geth.ipc 是rpc连接本地geth节点用的



## 示例

```
F:\work\src\github.com\ethereum\go-ethereum\cmd\geth>.\geth.exe --datadir "d:\gethtest" --testnet console
INFO [07-25|11:36:33.552] Maximum peer count                       ETH=25 LES=0 total=25
INFO [07-25|11:36:33.585] Starting peer-to-peer node               instance=Geth/v1.8.13-unstable/windows-amd64/go1.10.1
INFO [07-25|11:36:33.591] Allocated cache and file handles         database=d:\\gethtest\\geth\\chaindata cache=768 handles=1024
INFO [07-25|11:36:35.957] Persisted trie from memory database      nodes=355 size=51.89kB time=995.5µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B
INFO [07-25|11:36:35.964] Initialised chain configuration          config="{ChainID: 3 Homestead: 0 DAO: <nil> DAOSupport: true EIP150: 0 EIP155: 10 EIP158: 10 Byzantium: 1700000 Constantinople: <nil> Engine: ethash}"
INFO [07-25|11:36:35.973] Disk storage enabled for ethash caches   dir=d:\\gethtest\\geth\\ethash count=3
INFO [07-25|11:36:35.978] Disk storage enabled for ethash DAGs     dir=C:\\Users\\zxl\\AppData\\Ethash count=2
INFO [07-25|11:36:35.981] Initialising Ethereum protocol           versions="[63 62]" network=3
INFO [07-25|11:36:35.987] Loaded most recent local header          number=3705666 hash=101c8c…6bb2cf td=9024009794277963
INFO [07-25|11:36:35.992] Loaded most recent local full block      number=3705666 hash=101c8c…6bb2cf td=9024009794277963
INFO [07-25|11:36:35.997] Loaded most recent local fast block      number=3705666 hash=101c8c…6bb2cf td=9024009794277963
INFO [07-25|11:36:36.002] Loaded local transaction journal         transactions=0 dropped=0
INFO [07-25|11:36:36.008] Regenerated local transaction journal    transactions=0 accounts=0
WARN [07-25|11:36:36.011] Blockchain not empty, fast sync disabled
INFO [07-25|11:36:36.015] Starting P2P networking
INFO [07-25|11:36:38.426] UDP listener up                          self=enode://9ab7b572e5a7208726f954a100b747cb2bfef7be5e4f6abe379597e583911b9de3243412e4fbee405cfe4018188d7cfe588b2aaf4f88cac1c127a4e710b20edb@[::]:30303
INFO [07-25|11:36:41.087] RLPx listener up                         self=enode://9ab7b572e5a7208726f954a100b747cb2bfef7be5e4f6abe379597e583911b9de3243412e4fbee405cfe4018188d7cfe588b2aaf4f88cac1c127a4e710b20edb@[::]:30303
INFO [07-25|11:36:41.091] IPC endpoint opened                      url=\\\\.\\pipe\\geth.ipc
Welcome to the Geth JavaScript console!

instance: Geth/v1.8.13-unstable/windows-amd64/go1.10.1
 modules: admin:1.0 debug:1.0 eth:1.0 miner:1.0 net:1.0 personal:1.0 rpc:1.0 txpool:1.0 web3:1.0

>
```

```
F:\work\src\github.com\ethereum\go-ethereum\cmd\geth>.\geth.exe attach \\.\pipe\geth.ipc
Welcome to the Geth JavaScript console!

instance: Geth/v1.8.13-unstable/windows-amd64/go1.10.1
 modules: admin:1.0 debug:1.0 eth:1.0 miner:1.0 net:1.0 personal:1.0 rpc:1.0 txpool:1.0 web3:1.0

>
```
