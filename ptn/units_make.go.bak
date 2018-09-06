package ptn

import (
	"encoding/json"
	"fmt"
	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag"
	common2 "github.com/palletone/go-palletone/dag/common"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/dag/storage"
	"log"
	"time"
)

var jsongenesis string = `{
  "version": "0.6.0-alpha",
  "alias": "PTN",
  "tokenAmount": 100000000000000000,
  "tokenDecimal": 8,
  "decimal_unit": "",
  "chainId": 1,
  "tokenHolder": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
  "initialParameters": {
    "mediatorInterval": 5
  },
  "immutableChainParameters": {
    "MinMediatorCount": 11,
    "MinMediatorInterval": 1
  },
  "initialTimestamp": 1535963775,
  "initialActiveMediators": 21,
  "initialMediatorCandidates": [
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    },
    {
      "Address": "P1NcbtaDEAf7hRFem71k6A2qQDbZo59RQfo",
      "InitPartPub": "AV0a95Ex-pTGAAYXg277329ewkWDOOdkuo-Va1ogVnEQiCc-efOZlFMFWCYDhld8uxoRggzxRJJzPJ0r4hKKVKRZOW-aLJYuGilc5ONNi3riQBCoOry7cX65yLx9_yMENAnWI_fN6USJpDG2dJBTCyuY-N_GOZf9wD-2qhP5-lDj",
      "Node": "pnode://280d9c3b5b0f43d593038987dc03edea62662ba5a9fecea0a1b216c0e0e6f59599896d4d3621f70fbbc63e05c95151e154c84aad7825008b118824a99d27541b@127.0.0.1:30303"
    }
  ],
  "systemConfig": {
    "depositRate": 0.02
  }
}`
var genesis = &core.Genesis{}

func makegenesis(memdb ptndb.Database) {
	err := json.Unmarshal([]byte(jsongenesis), genesis)
	if err != nil {
		log.Fatal("err", err.Error())
	}
	fmt.Printf("=======%#v\n", genesis)
	gp := modules.InitGlobalProp(genesis)
	storage.StoreGlobalProp(memdb, gp)
}
func MakeDags(Memdb ptndb.Database, unitAccount int) (*dag.Dag, error) {
	dag, _ := dag.NewDagForTest(Memdb)
	header := NewHeader([]common.Hash{}, []modules.IDType16{modules.PTNCOIN}, []byte{})
	header.Number.AssetID = modules.PTNCOIN
	header.Number.IsMain = true
	header.Number.Index = 0
	header.Authors = &modules.Authentifier{"", []byte{}, []byte{}, []byte{}}
	header.Witness = []*modules.Authentifier{&modules.Authentifier{"", []byte{}, []byte{}, []byte{}}}
	tx, _ := NewCoinbaseTransaction()
	txs := modules.Transactions{tx}
	genesisUnit := NewUnit(header, txs)
	//fmt.Printf("--------这是最新块----unit.UnitHeader-----%#v\n", genesisUnit.UnitHeader)
	err := SaveGenesis(dag.Db, genesisUnit)
	if err != nil {
		fmt.Println("SaveGenesis, err", err)
		return nil, err
	}
	//fmt.Printf("--------这是最新块----unit-----%#v\n", genesisUnit)
	//fmt.Printf("--------这是最新块----unit.UnitHeader-----%#v\n", genesisUnit.UnitHeader)
	//fmt.Printf("--------这是最新块----unit.Txs-----%#v\n", genesisUnit.Txs[0].Hash())
	//fmt.Printf("--------这是最新块----unit.UnitHash-----%#v\n", genesisUnit.UnitHash)
	//fmt.Printf("--------这是最新块----unit.UnitHeader.ParentsHash-----%#v\n", genesisUnit.UnitHeader.ParentsHash)
	//fmt.Printf("--------这是最新块----unit.UnitHeader.Number.Index-----%#v\n", genesisUnit.UnitHeader.Number.Index)
	units,_ := newDag(dag.Db, genesisUnit, unitAccount)
	fmt.Println("len(units).........", len(units))
	//for i, v := range units {
	//	fmt.Printf("%d====%#v\n", i, v)
	//}
	uu := dag.CurrentUnit()
	fmt.Printf("current===>>>%#v\n", uu)
	fmt.Printf("--------current----unit.UnitHeader-----%#v\n", uu.UnitHeader)
	//fmt.Printf("--------这是最新块----unit.Txs-----%#v\n", uu.Txs[0].Hash())
	fmt.Printf("--------current----unit.UnitHash-----%#v\n", uu.UnitHash)
	fmt.Printf("--------current----unit.UnitHeader.ParentsHash-----%#v\n", uu.UnitHeader.ParentsHash)
	fmt.Printf("--------current----unit.UnitHeader.Number.Index-----%#v\n", uu.UnitHeader.Number.Index)
	fmt.Println("---------------进入循坏-------------")
	for {
		fmt.Printf("查找===>>>%#v\n", uu)
		fmt.Printf("--------current----unit.UnitHeader-----%#v\n", uu.UnitHeader)
		fmt.Printf("--------current----unit.UnitHash-----%#v\n", uu.UnitHash)
		fmt.Printf("--------current----unit.UnitHeader.ParentsHash-----%#v\n", uu.UnitHeader.ParentsHash)
		fmt.Printf("--------current----unit.UnitHeader.Number.Index-----%#v\n", uu.UnitHeader.Number.Index)
		if len(uu.ParentHash()) > 0 {
			uu = dag.GetUnit(uu.ParentHash()[0])
		} else {
			break
		}
	}
	fmt.Println("MakeDags=",dag.GetUnitByNumber(index))
	uu := dag.CurrentUnit()
		fmt.Printf("current===>>>%#v\n",uu)
		//fmt.Printf("--------current----unit.UnitHeader-----%#v\n", uu.UnitHeader)
		////fmt.Printf("--------这是最新块----unit.Txs-----%#v\n", uu.Txs[0].Hash())
		//fmt.Printf("--------current----unit.UnitHash-----%#v\n", uu.UnitHash)
		//fmt.Printf("--------current----unit.UnitHeader.ParentsHash-----%#v\n", uu.UnitHeader.ParentsHash)
		//fmt.Printf("--------current----unit.UnitHeader.Number.Index-----%#v\n", uu.UnitHeader.Number.Index)
	//if uu == nil {
	//	fmt.Println("dag.CurrentUnit()====>get nil===>",uu)
	//}else{
	//	//fmt.Printf("current===>>>%#v\n",uu)
	//	//fmt.Printf("--------current----unit.UnitHeader-----%#v\n", uu.UnitHeader)
	//	////fmt.Printf("--------这是最新块----unit.Txs-----%#v\n", uu.Txs[0].Hash())
	//	//fmt.Printf("--------current----unit.UnitHash-----%#v\n", uu.UnitHash)
	//	//fmt.Printf("--------current----unit.UnitHeader.ParentsHash-----%#v\n", uu.UnitHeader.ParentsHash)
	//	//fmt.Printf("--------current----unit.UnitHeader.Number.Index-----%#v\n", uu.UnitHeader.Number.Index)
	//	fmt.Println("---------------进入循坏-------------")
	//	for {
	//		fmt.Printf("查找===>>>%#v\n",uu)
	//		fmt.Printf("--------current----unit.UnitHeader-----%#v\n", uu.UnitHeader)
	//		fmt.Printf("--------current----unit.UnitHash-----%#v\n", uu.UnitHash)
	//		fmt.Printf("--------current----unit.UnitHeader.ParentsHash-----%#v\n", uu.UnitHeader.ParentsHash)
	//		fmt.Printf("--------current----unit.UnitHeader.Number.Index-----%#v\n", uu.UnitHeader.Number.Index)
	//		if len(uu.ParentHash()) > 0 {
	//			uu = dag.GetUnit(uu.ParentHash()[0])
	//		}else{
	//			break
	//		}
	//	}
	//	fmt.Println("---------------退出循坏-------------")
	//}
	return dag, nil
}
func newDag(memdb ptndb.Database, gunit *modules.Unit, number int) (modules.Units, error) {
	units := make(modules.Units, number)
	par := gunit
	//fmt.Println("len(units).........",len(units))
	//fmt.Println("number.........",number)
	for i := 0; i < number; i++ {
		//fmt.Println("createUnit",i)
		header := NewHeader([]common.Hash{par.UnitHash}, []modules.IDType16{modules.PTNCOIN}, []byte{})
		header.Number.AssetID = par.UnitHeader.Number.AssetID
		header.Number.IsMain = par.UnitHeader.Number.IsMain
		header.Number.Index = par.UnitHeader.Number.Index + 1
		header.Authors = &modules.Authentifier{"", []byte{}, []byte{}, []byte{}}
		header.Witness = []*modules.Authentifier{&modules.Authentifier{"", []byte{}, []byte{}, []byte{}}}
		tx, _ := NewCoinbaseTransaction()
		txs := modules.Transactions{tx}
		unit := NewUnit(header, txs)
		//fmt.Println("start saveUnit")
		err := SaveUnit(memdb, unit, true)
		if err != nil {
			fmt.Println("Save==", err)
		}
		//fmt.Printf("--------这是父块----unit-----%#v\n",unit)
		//fmt.Printf("--------这是父块----unit.UnitHeader-----%#v\n",unit.UnitHeader)
		//fmt.Printf("--------这是父块----unit.Txs-----%#v\n", unit.Txs[0].Hash())
		//fmt.Printf("--------这是父块----unit.UnitHash-----%#v\n",unit.UnitHash)
		//fmt.Printf("--------这是父块----unit.UnitHeader.ParentsHash-----%#v\n",unit.UnitHeader.ParentsHash)
		//fmt.Printf("--------这是父块----unit.UnitHeader.Number.Index-----%#v\n",unit.UnitHeader.Number.Index)
		//fmt.Println("createUnit",i)
		units[i] = unit
		par = unit
	}
	return units, nil
}
func SaveGenesis(db ptndb.Database, unit *modules.Unit) error {
	//fmt.Println("unit.NumberU64()====",unit.NumberU64())
	if unit.NumberU64() != 0 {
		return fmt.Errorf("can't commit genesis unit with number > 0")
	}
	//fmt.Println("start saveUnit")
	err := SaveUnit(db, unit, true)
	if err != nil {
		fmt.Println("SaveGenesis==", err)
	}
	//fmt.Println("end saveUnit")
	return nil
}

func SaveUnit(db ptndb.Database, unit *modules.Unit, isGenesis bool) error {
	if unit.UnitSize == 0 || unit.Size() == 0 {
		log.Println("Unit is null")
		//return fmt.Errorf("Unit is null")
	}
	// step1. check unit signature, should be compare to mediator list

	errno := common2.ValidateUnitSignature(db, unit.UnitHeader, isGenesis)
	if int(errno) != modules.UNIT_STATE_VALIDATED && int(errno) != modules.UNIT_STATE_AUTHOR_SIGNATURE_PASSED {
		return fmt.Errorf("Validate unit signature error, errno=%d", errno)
	}
	// step2. check unit size
	if unit.UnitSize != unit.Size() {
		log.Println("Validate size", "error", "Size is invalid")
		//return modules.ErrUnit(-1)
	}
	// step3. check transactions in unit
	_, isSuccess, err := common2.ValidateTransactions(db, &unit.Txs, isGenesis)
	if isSuccess != true {
		fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
		//return fmt.Errorf("Validate unit(%s) transactions failed: %v", unit.UnitHash.String(), err)
	}
	// step4. save unit header
	// key is like "[HEADER_PREFIX][chain index number]_[chain index]_[unit hash]"
	if err := storage.SaveHeader(db, unit.UnitHash, unit.UnitHeader); err != nil {
		log.Println("SaveHeader:", "error", err.Error())
		//return modules.ErrUnit(-3)
	}
	// step5. save unit hash and chain index relation
	// key is like "[UNIT_HASH_NUMBER][unit_hash]"

	//fmt.Printf("==============unit.UnitHeader.Number=%#v\n",unit.UnitHeader.Number)
	//fmt.Printf("--------这是最新块----unit.UnitHash-----%#v\n", unit.UnitHash)
	if err := storage.SaveNumberByHash(db,unit.UnitHash, unit.UnitHeader.Number); err != nil {

		log.Println("SaveHashNumber:", "error", err.Error())
		//return fmt.Errorf("Save unit hash and number error")
		//fmt.Println("jinru===============")
	}

	if err := storage.SaveHashByNumber(db,unit.UnitHash, unit.UnitHeader.Number); err != nil {
		log.Println("SaveNumberByHash:", "error", err.Error())
		//return fmt.Errorf("Save unit hash and number error")
		//fmt.Println("jinru===============")

	if err := storage.SaveTxLookupEntry(db, unit); err != nil {
		return err

	}
	//if err := storage.SaveTxLookupEntry(db,unit); err != nil {
	//	return err
	//}
	// update state
	storage.PutCanonicalHash(db, unit.UnitHash, unit.NumberU64())
	storage.PutHeadHeaderHash(db, unit.UnitHash)
	storage.PutHeadUnitHash(db, unit.UnitHash)
	storage.PutHeadFastUnitHash(db, unit.UnitHash)
	// todo send message to transaction pool to delete unit's transactions
	return nil
}
func NewUnit(header *modules.Header, txs modules.Transactions) *modules.Unit {
	u := &modules.Unit{
		UnitHeader: header,
		Txs:        txs,
	}
	u.ReceivedAt = time.Now()
	u.UnitSize = u.Size()
	u.UnitHash = u.Hash()
	return u
}
func NewHeader(parents []common.Hash, asset []modules.IDType16, extra []byte) *modules.Header {
	hashs := make([]common.Hash, 0)
	hashs = append(hashs, parents...) // 切片指针传递的问题，这里得再review一下。
	var b []byte
	return &modules.Header{ParentsHash: hashs, AssetIDs: asset, Extra: append(b, extra...), Creationdate: time.Now().Unix()}
}
func NewCoinbaseTransaction() (*modules.Transaction, error) {
	input := &modules.Input{}
	output := &modules.Output{}
	payload := modules.PaymentPayload{
		Input:  []*modules.Input{input},
		Output: []*modules.Output{output},
	}
	msg := modules.Message{
		App:     modules.APP_PAYMENT,
		Payload: payload,
	}
	coinbase := &modules.Transaction{
		TxMessages: []modules.Message{msg},
	}
	coinbase.TxHash = coinbase.Hash()
	return coinbase, nil
}
