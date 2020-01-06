## 准备

1.安装go

## 下载源码、编译


1.下载 btcd 代码

```
mkdir F:\work\src\golang.org\x
cd F:\work\src\golang.org\x
git clone https://github.com/golang/crypto.git
go get -u github.com/btcsuite/btcd
```

2.编译出 btcd 可执行程序

```
cd F:\work\src\github.com\btcsuite\btcd
go build
```



## 启动

+ 启动 btcd

```
cd F:\work\src\github.com\btcsuite\btcd
.\btcd.exe -u test -P 123456 --datadir "d:\\btctest\\" --testnet --txindex --addrindex
```

-u 是 rpcuser ，-P 是 rpcpasswd ， --datadir 是区块数据目录， --testnet 是指定测试链
去掉参数 --testnet 即是正式链
--txindex 是建立交易索引， --addrindex 是建立地址索引，两者顺序不能颠倒



## 示例

```
F:\work\src\github.com\btcsuite\btcd>.\btcd.exe -u test -P 123456 --datadir "d:\\btctest\\" --testnet
2018-06-27 13:55:31.585 [INF] BTCD: Version 0.12.0-beta
2018-06-27 13:55:31.610 [INF] BTCD: Loading block database from 'd:\btctest\testnet\blocks_ffldb'
2018-06-27 13:55:31.659 [INF] BTCD: Block database loaded
2018-06-27 13:55:31.686 [INF] INDX: cf index is enabled
2018-06-27 13:55:31.688 [INF] INDX: Catching up indexes from height -1 to 0
2018-06-27 13:55:31.689 [INF] INDX: Indexes caught up to height 0
2018-06-27 13:55:31.689 [INF] CHAN: Chain state (height 0, hash 000000000933ea01ad0ee984209779baaec3ced90fa3f408719526f8d77f4943, totaltx 1, work 4295032833)
2018-06-27 13:55:31.700 [INF] RPCS: RPC server listening on [::1]:18334
2018-06-27 13:55:31.700 [INF] RPCS: RPC server listening on 127.0.0.1:18334
2018-06-27 13:55:31.700 [INF] AMGR: Loaded 0 addresses from file 'd:\btctest\testnet\peers.json'
2018-06-27 13:55:31.701 [INF] CMGR: Server listening on 0.0.0.0:18333
2018-06-27 13:55:31.701 [INF] CMGR: Server listening on [::]:18333
2018-06-27 13:55:31.739 [INF] CMGR: 1 addresses found from DNS seed testnet-seed.bluematt.me
2018-06-27 13:55:31.739 [INF] CMGR: 25 addresses found from DNS seed testnet-seed.bitcoin.schildbach.de
2018-06-27 13:55:31.740 [INF] CMGR: 23 addresses found from DNS seed testnet-seed.bitcoin.jonasschnelli.ch
2018-06-27 13:55:31.742 [INF] CMGR: 24 addresses found from DNS seed seed.tbtc.petertodd.org
2018-06-27 13:55:36.964 [INF] SYNC: New valid peer 13.78.14.162:18333 (outbound) (/Satoshi:0.16.0/)
2018-06-27 13:55:36.965 [INF] SYNC: Syncing to block height 1326466 from peer 13.78.14.162:18333
2018-06-27 13:55:36.968 [INF] SYNC: Downloading headers for blocks 1 to 546 from peer 13.78.14.162:18333
2018-06-27 13:55:37.030 [INF] SYNC: New valid peer 172.105.194.235:18333 (outbound) (/Satoshi:0.16.0(bitcore)/)
```

