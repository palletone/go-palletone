package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/palletone/go-palletone/contracts/manger"
	"github.com/palletone/go-palletone/core/vmContractPub/crypto"
	"github.com/palletone/go-palletone/core/vmContractPub/util"

	"encoding/hex"

	"github.com/spf13/viper"

	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/storage"
)

func computeProposalTxID(nonce, creator []byte) (string, error) {
	opdata := append(nonce, creator...)
	digest := util.ComputeSHA256(opdata)
	return hex.EncodeToString(digest), nil
}

func putval(args *[]string) [][]byte {
	if len(*args) < 4 {
		//		//putval example
		//		chaincodeFunc := "putval"
		//		testKey := "testk"
		//		testValue := "testvNew"
		//		argBytes := util.ToChaincodeArgs(chaincodeFunc, testKey, testValue)
		//		return argBytes
		fmt.Println("Params : putval, testKey, testValue")
		return nil
	} else {
		//putval
		chaincodeFunc := (*args)[1]
		testKey := (*args)[2]
		testValue := (*args)[3]
		argBytes := util.ToChaincodeArgs(chaincodeFunc, testKey, testValue)
		return argBytes
	}
}

func getval(args *[]string) [][]byte {
	if len(*args) < 3 {
		//		//getval example
		//		chaincodeFunc := "getval"
		//		testKey := "testk"
		//		argBytes := util.ToChaincodeArgs(chaincodeFunc, testKey)
		//		return argBytes
		fmt.Println("Params : getval, testKey")
		return nil
	} else {
		//getval
		chaincodeFunc := (*args)[1]
		testKey := (*args)[2]
		argBytes := util.ToChaincodeArgs(chaincodeFunc, testKey)
		return argBytes
	}
}

func multiSigAddrBTC(args *[]string) [][]byte {
	if len(*args) < 4 {
		//		//MultiAddr example
		//		chaincodeFunc := "multiSigAddrBTC"
		//		pubkeyAlice := "029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef5"
		//		pubkeyBob := "020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb"
		//		argBytes := util.ToChaincodeArgs(chaincodeFunc, pubkeyAlice, pubkeyBob)
		//		return argBytes
		fmt.Println("Params : multiSigAddrBTC, pubkeyAlice, pubkeyBob")
		return nil
	} else {
		//MultiAddr
		chaincodeFunc := (*args)[1]
		pubkeyAlice := (*args)[2]
		pubkeyBob := (*args)[3]
		argBytes := util.ToChaincodeArgs(chaincodeFunc, pubkeyAlice, pubkeyBob)
		return argBytes
	}
}

func withdrawBTC(args *[]string) [][]byte {
	if len(*args) < 3 {
		//		//withdrawBTC example
		//		chaincodeFunc := "withdrawBTC"
		//		transactionhex := "010000000236045404e65bd741109db92227ca0dc9274ef717a6612c96cd77b24a17d1bcd700000000b400473044022024e6a6ca006f25ccd3ebf5dadf21397a6d7266536cd336061cd17cff189d95e402205af143f6726d75ac77bc8c80edcb6c56579053d2aa31601b23bc8da41385dd86014c69522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553aeffffffff7c1f7d5407b41abf29d41cf6f122ef2d40f76d956900d2c89314970951ef5b9400000000b40047304402206a1d7a2ae07840957bee708b6d3e1fbe7858760ac378b1e21209b348c1e2a5c402204255cd4cd4e5b5805d44bbebe7464aa021377dca5fc6bf4a5632eb2d8bc9f9e4014c69522103940ab29fbf214da2d8ec99c47db63879957311bd90d2f1c635828604d541051421020106ca23b4f28dbc83838ee4745accf90e5621fe70df5b1ee8f7e1b3b41b64cb21029d80ff37838e4989a6aa26af41149d4f671976329e9ddb9b78fdea9814ae6ef553aeffffffff014431d309000000001976a914bddc9a62e9b7c3cfdbe1c817520e24e32c339f3288ac00000000"
		//		argBytes := util.ToChaincodeArgs(chaincodeFunc, transactionhex)
		//		return argBytes
		fmt.Println("Params : withdrawBTC, transactionhex")
		return nil
	} else {
		//MultiAddr
		chaincodeFunc := (*args)[1]
		transactionhex := (*args)[2]
		argBytes := util.ToChaincodeArgs(chaincodeFunc, transactionhex)
		return argBytes
	}
}

func multiSigAddrETH(args *[]string) [][]byte {
	if len(*args) < 4 {
		//		//multiSigAddrETH example
		//		chaincodeFunc := "multiSigAddrETH"
		//		addrAlice := "0x7d7116a8706ae08baa7f4909e26728fa7a5f0365"
		//		addrBob := "0xaAA919a7c465be9b053673C567D73Be860317963"
		//		argBytes := util.ToChaincodeArgs(chaincodeFunc, addrAlice, addrBob)
		//		return argBytes
		fmt.Println("Params : multiSigAddrETH, addrAlice, addrBob")
		return nil

	} else {
		//multiSigAddrETH
		chaincodeFunc := (*args)[1]
		addrAlice := (*args)[2]
		addrBob := (*args)[3]
		argBytes := util.ToChaincodeArgs(chaincodeFunc, addrAlice, addrBob)
		return argBytes
	}
}

func calSigETH(args *[]string) [][]byte {
	if len(*args) < 3 {
		//		//calSigETH example
		//		chaincodeFunc := "calSigETH"
		//		addrAlice := "0xaAA919a7c465be9b053673C567D73Be860317963"
		//		amountWei := "1000000000000000000"
		//		argBytes := util.ToChaincodeArgs(chaincodeFunc, addrAlice, amountWei)
		//		return argBytes
		fmt.Println("Params : calSigETH, addrAlice, amountWei")
		return nil
	} else {
		//calSigETH
		chaincodeFunc := (*args)[1]
		addrAlice := (*args)[2]
		amountWei := (*args)[3]
		argBytes := util.ToChaincodeArgs(chaincodeFunc, addrAlice, amountWei)
		return argBytes
	}
}

func multMoreSys(args [][]byte) {
	fmt.Println("mult enter..................")
	chainID := util.GetTestChainID()

	var wg sync.WaitGroup

	var tmout time.Duration = 500 * time.Second
	var txid string = "1234567890" //default

	nonce, err := crypto.GetRandomNonce()
	if err == nil {
		txid, err = computeProposalTxID(nonce, []byte("glh"))
	}
	fmt.Println("++++++++++++++++ txid:" + txid)

	wg.Add(1)
	go func(timeout time.Duration, txid string) {
		unit, err := manger.Invoke(chainID, []byte{0x95, 0x27}, txid, args, timeout)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			if unit != nil {
				fmt.Println("len(unit.WriteSet) ==== ==== ", len(unit.WriteSet))
				for k, v := range unit.WriteSet {
					fmt.Printf("k[%d], v[%v]\n", k, v)
				}
			} else {
				fmt.Println("Not nil error. But nil unit !!!")
			}
		}
		wg.Done()
	}(tmout, txid)
	wg.Wait()
}

func TestExecSysCCMult(args *[]string) {
	//
	for _, arg := range *args {
		fmt.Println(arg)
	}

	//
	var argBytes [][]byte
	if len(*args) < 2 {
		helper()
		return
	} else {
		chaincodeFunc := (*args)[1]
		switch chaincodeFunc {
		case "multiSigAddrBTC":
			argBytes = multiSigAddrBTC(args)
		case "withdrawBTC":
			argBytes = withdrawBTC(args)

		case "multiSigAddrETH":
			argBytes = multiSigAddrETH(args)
		case "calSigETH":
			argBytes = calSigETH(args)

		case "putval":
			argBytes = putval(args)
		case "getval":
			argBytes = getval(args)
		default:
			helper()
			return
		}
	}

	viper.Set("peer.fileSystemPath", "d:\\chaincodes")
	viper.Set("peer.address", "127.0.0.1:12345")
	viper.Set("chaincode.executetimeout", 20*time.Second)

	// System chaincode has to be enabled
	viper.Set("chaincode.system", map[string]string{"sample_syscc": "true"})

	//	//
	//	manger.Init(nil)
	dag, err := creatTestDag()
	if err == nil {
		manger.Init(dag)
	} else {
		fmt.Println("creatTestDag err")
		return
	}

	//
	if argBytes != nil {
		multMoreSys(argBytes)
	}
}
func creatTestDag() (*dag.Dag, error) {
	path := "D:\\test\\levedb"

	dagconfig.DbPath = path
	db, err := storage.Init(path, 16, 16)
	if err != nil {
		return nil, err
	}
	dag, err := dag.NewDag(db)
	if err != nil {
		return nil, err
	}

	return dag, nil
}
func helper() {
	fmt.Println("functions : putval, getval, multiSigAddrBTC, withdrawBTC, multiSigAddrETH, calSigETH")
	fmt.Println("Params : putval, testKey, testValue")
	fmt.Println("Params : getval, testKey")
	fmt.Println("Params : multiSigAddrBTC, pubkeyAlice, pubkeyBob")
	fmt.Println("Params : withdrawBTC, transactionhex")
	fmt.Println("Params : multiSigAddrETH, addrAlice, addrBob")
	fmt.Println("Params : calSigETH, addrAlice, amountWei")
}
func main() {
	args := os.Args
	TestExecSysCCMult(&args)
}
