## Go palletone

Official golang implementation of the palletone protocol.

[![API Reference](
https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
)](https://godoc.org/github.com/palletone/go-palletone)
[![Go Report Card](https://goreportcard.com/badge/github.com/palletone/go-palletone)](https://goreportcard.com/report/github.com/palletone/go-palletone)
[![Travis](https://travis-ci.org/palletone/go-palletone.svg?branch=master)](https://travis-ci.org/palletone/go-palletone)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/palletone/go-palletone?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)


## Building the source

For prerequisites and detailed build instructions please read the
[Installation Instructions](https://github.com/palletone/go-palletone/wiki/Building-palletone)
on the wiki.

Building gptn requires both a Go (version 1.7 or later) and a C compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run

    make gptn

or, to build the full suite of utilities:

    make all

## Executables

The go-palletone project comes with several wrappers/executables found in the `cmd` directory.

| Command    | Description |
|:----------:|-------------|
| **`gptn`** | Our main palletone CLI client. It is the entry point into the palletone network (main-, test- or private net), capable of running as a full node (default) archive node (retaining all historical state) or a light node (retrieving data live). It can be used by other processes as a gateway into the palletone network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. `gptn --help` and the [CLI Wiki page](https://github.com/palletone/go-palletone/wiki/Command-Line-Options) for command line options. |

## Running gptn

Going through all the possible command line flags is out of scope here (please consult our
[CLI Wiki page](https://github.com/palletone/go-palletone/wiki/Command-Line-Options)), but we've
enumerated a few common parameter combos to get you up to speed quickly on how you can run your
own Geth instance.

### Full node on the main palletone network

By far the most common scenario is people wanting to simply interact with the palletone network:
create accounts; transfer funds; deploy and interact with contracts. For this particular use-case
the user doesn't care about years-old historical data, so we can fast-sync quickly to the current
state of the network. To do so:

```
$ gptn --config /path/to/your_config.toml console 
```

This command will:

 * Start gptn in fast sync mode (default, can be changed with the `--syncmode` flag), causing it to
   download more data in exchange for avoiding processing the entire history of the palletone network,
   which is very CPU intensive.
 * Start up Geth's built-in interactive [JavaScript console](https://github.com/palletone/go-palletone/wiki/JavaScript-Console),
   (via the trailing `console` subcommand) through which you can invoke all official [`web3` methods](https://github.com/palletone/wiki/wiki/JavaScript-API)
   as well as Geth's own [management APIs](https://github.com/palletone/go-palletone/wiki/Management-APIs).
   This too is optional and if you leave it out you can always attach to an already running Geth instance
   with `gptn attach`.

### Full node on the palletone test network

Transitioning towards developers, if you'd like to play around with creating palletone contracts, you
almost certainly would like to do that without any real money involved until you get the hang of the
entire system. In other words, instead of attaching to the main network, you want to join the **test**
network with your node, which is fully equivalent to the main network, but with play-Ether only.

```
$ gptn --config /path/to/your_config.toml --testnet console
```

The `console` subcommand have the exact same meaning as above and they are equally useful on the
testnet too. Please see above for their explanations if you've skipped to here.

Specifying the `--testnet` flag however will reconfigure your Geth instance a bit:

 * Instead of using the default data directory (`~/.palletone` on Linux for example), Geth will nest
   itself one level deeper into a `testnet` subfolder (`~/.palletone/testnet` on Linux). Note, on OSX
   and Linux this also means that attaching to a running testnet node requires the use of a custom
   endpoint since `gptn attach` will try to attach to a production node endpoint by default. E.g.
   `gptn attach <datadir>/testnet/gptn.ipc`. Windows users are not affected by this.
 * Instead of connecting the main palletone network, the client will connect to the test network,
   which uses different P2P bootnodes, different network IDs and genesis states.
   
*Note: Although there are some internal protective measures to prevent transactions from crossing
over between the main network and test network, you should make sure to always use separate accounts
for play-money and real-money. Unless you manually move accounts, Geth will by default correctly
separate the two networks and will not make any accounts available between them.*


### Configuration

As an alternative to passing the numerous flags to the `gptn` binary, you can also pass a configuration file via:

```
$ gptn --config /path/to/your_config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to export your existing configuration:

```
$ gptn --your-favourite-flags dumpconfig
```

e.g. call it palletone.toml:

```
[Consensus]
Engine="solo"

[Log]
OutputPaths =["stdout","./log/all.log"]
ErrorOutputPaths= ["stderr","./log/error.log"]
LoggerLvl="info"   # ("debug", "info", "warn","error", "dpanic", "panic", and "fatal")
Encoding="console" # console,json
Development =true

[Dag]
DbPath="./leveldb"
DbName="palletone.db"

[Ada]
Ada1="ada1_config"
Ada2="ada2_config"

[Node]
DataDir = "./data1"
KeyStoreDir="./data1/keystore"
IPCPath = "./data1/gptn.ipc"
HTTPPort = 8541
HTTPVirtualHosts = ["0.0.0.0"]
HTTPCors = ["*"]

[Ptn]
NetworkId = 3369

[P2P]
ListenAddr = "0.0.0.0:30301"
#BootstrapNodes = ["pnode://228f7e50031457d804ce6021f4a211721bacb9abba9585870efea55780bb744005a7f22e22938040684cdec32c748968f5dbe19822d4fbb44c6aaa69e7abdfee@127.0.0.1:30301"]
```


### Operating a private network

Maintaining your own private network is more involved as a lot of configurations taken for granted in
the official networks need to be manually set up.

#### Defining the private genesis state

First, you'll need to create the genesis state of your networks, which all nodes need to be aware of
and agree upon. This consists of a small JSON file (e.g. call it `genesis.json`):

```json
{
  "height": "0",
  "version": "0.6.0",
  "tokenAmount": 1000000000,
  "tokenDecimal": 8,
  "chainId": 0,
  "tokenHolder": "P1Kp2hcLhGEP45Xgx7vmSrE37QXunJUd8gJ",
  "systemConfig": {
    "mediatorSlot": 5,
    "mediatorCount": 21,
    "mediatorList": [
      "dfba98bb5c52bba028e2cc487cbd1084"
    ],
    "mediatorCycle": 86400,  <!--24 Hours-->
    "depositRate": 0.02
  }
}
```

With the genesis state defined in the above JSON file, you'll need to initialize **every** Geth node
with it prior to starting it up to ensure all blockchain parameters are correctly set:

```
$ gptn init path/to/genesis.json
```

#### Starting up your member nodes

With the bootnode operational and externally reachable (you can try `telnet <ip> <port>` to ensure
it's indeed reachable), start every subsequent Geth node pointed to the bootnode for peer discovery
via the `--bootnodes` flag. It will probably also be desirable to keep the data directory of your
private network separated, so do also specify a custom `--datadir` flag.

```
$ gptn --datadir=path/to/custom/data/folder
```

*Note: Since your network will be completely cut off from the main and test networks, you'll also
need to configure a miner to process transactions and create new blocks for you.*

## Contribution

Thank you for considering to help out with the source code! We welcome contributions from
anyone on the internet, and are grateful for even the smallest of fixes!

If you'd like to contribute to go-palletone, please fork, fix, commit and send a pull request
for the maintainers to review and merge into the main code base. If you wish to submit more
complex changes though, please check up with the core devs first on [our gitter channel](https://gitter.im/palletone/go-palletone)
to ensure those changes are in line with the general philosophy of the project and/or get some
early feedback which can make both your efforts much lighter as well as our review and merge
procedures quick and simple.

Please make sure your contributions adhere to our coding guidelines:

 * Code must adhere to the official Go [formatting](https://golang.org/doc/effective_go.html#formatting) guidelines (i.e. uses [gofmt](https://golang.org/cmd/gofmt/)).
 * Code must be documented adhering to the official Go [commentary](https://golang.org/doc/effective_go.html#commentary) guidelines.
 * Pull requests need to be based on and opened against the `master` branch.
 * Commit messages should be prefixed with the package(s) they modify.
   * E.g. "eth, rpc: make trace configs optional"

Please see the [Developers' Guide](https://github.com/palletone/go-palletone/wiki/Developers'-Guide)
for more details on configuring your environment, managing project dependencies and testing procedures.

## License

The go-palletone binaries (i.e. all code inside of the `cmd` directory) is licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also included
in our repository in the `COPYING` file.
