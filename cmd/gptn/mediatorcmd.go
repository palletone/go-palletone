/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"gopkg.in/urfave/cli.v1"
)

var (
	// append by Albert·Gou
	nodeInfoCommand = cli.Command{
		Action:    utils.MigrateFlags(showNodeInfo),
		Name:      "nodeInfo",
		Usage:     "get info of current node",
		ArgsUsage: "",
		Category:  "MEDIATOR COMMANDS",
		Description: `
The output of this command will be used to set the genesis json file.
`,
	}

	// append by Albert·Gou
	timestampCommand = cli.Command{
		Action:    utils.MigrateFlags(getTimestamp),
		Name:      "timestamp",
		Usage:     "get the timestamp of the Unix epoch at the specified time",
		ArgsUsage: "<specified time>",
		Flags: []cli.Flag{
			timeStringFlag,
		},
		Category: "MEDIATOR COMMANDS",
		Description: `
The format of the specified time should be like "2006-01-02 15:04:05", 
and If not specified, displays the timestamp of 
the time when the current command is running.
`,
	}

	timeStringFlag = cli.StringFlag{
		Name:  "timeString",
		Usage: "time formatted as \"2006-01-02 15:04:05\"",
		Value: "",
	}

	mediatorCommand = cli.Command{
		Name:      "mediator",
		Usage:     "Manage mediators",
		ArgsUsage: "",
		Category:  "MEDIATOR COMMANDS",
		Description: `
    Manage mediators, list all existing mediators, create a new mediator.
`,
		Subcommands: []cli.Command{
			// 创建Mediator初始秘钥分片
			{
				Action:    utils.MigrateFlags(createInitDKS),
				Name:      "initdks",
				Usage:     "Generate the initial distributed key share.",
				ArgsUsage: "",
				Category:  "MEDIATOR COMMANDS",
				Description: `
The output of this command will be used to initialize the DistKeyGenerator.
`,
			},

			// 列出当前区块链所有mediator的地址
			{
				Action:    utils.MigrateFlags(listMediators),
				Name:      "list",
				Usage:     "List all mediators.",
				ArgsUsage: "",
				Category:  "MEDIATOR COMMANDS",
				Description: `
List all existing mediator addresses.
`,
			},
		},
	}
)

// author Albert·Gou
func createInitDKS(ctx *cli.Context) error {
	secStr, pubStr := core.CreateInitDKS()

	fmt.Println("Generate a initial distributed key share:")
	fmt.Println("{")
	fmt.Println("\tprivate key: ", secStr)
	fmt.Println("\tpublic key: ", pubStr)
	fmt.Println("}")

	return nil
}

// author Albert·Gou
func getNodeInfo(ctx *cli.Context) string {
	_, cfg := makeConfigNode(ctx, false)
	privateKey := cfg.Node.NodeKey()
	listenAddr := cfg.P2P.ListenAddr
	if strings.HasPrefix(listenAddr, ":") {
		listenAddr = "127.0.0.1" + listenAddr
	}

	nodeID := discover.PubkeyID(&privateKey.PublicKey)
	nodeInfo := "pnode://" + nodeID.String() + "@" + listenAddr

	return nodeInfo
}

// author Albert·Gou
func showNodeInfo(ctx *cli.Context) error {
	nodeStr := getNodeInfo(ctx)

	fmt.Println(nodeStr)

	return nil
}

// author Albert·Gou
func getTimestamp(ctx *cli.Context) error {
	var timeUnix time.Time
	var err error

	timeStr := ctx.Args().First()
	if len(timeStr) == 0 {
		timeUnix = time.Now()
	} else {
		timeUnix, err = time.Parse("2006-01-02 15:04:05", timeStr)
		if err != nil {
			fmt.Println("Please enter the time in the following format: \"2006-01-02 15:04:05\"")
			return nil
		}
	}

	fmt.Println(timeUnix.Unix())

	return nil
}

func listMediators(ctx *cli.Context) error {
	node := makeFullNode(ctx, false)

	Dbconn, err := node.OpenDatabase(dagconfig.DagConfig.DbPath, 0, 0)
	if err != nil {
		fmt.Println("leveldb init failed!")
		return err
	}

	dag, err := dag.NewDag(Dbconn, node.CacheDb, false)
	if err != nil {
		fmt.Println("leveldb init failed!")
		return err
	}

	mas := dag.GetMediators()
	fmt.Println("\nList all existing mediator addresses:")
	fmt.Println("[")

	for address := range mas {
		fmt.Printf("\t%s,\n", address.Str())
	}

	fmt.Println("]")

	return nil
}
