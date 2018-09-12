// Copyright 2018 PalletOne

// Copyright 2016 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/dedis/kyber/pairing/bn256"
	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/configure"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/ptn"
	"gopkg.in/urfave/cli.v1"
)

var (
	makecacheCommand = cli.Command{
		Action:    utils.MigrateFlags(makecache),
		Name:      "makecache",
		Usage:     "Generate ethash verification cache (for testing)",
		ArgsUsage: "<blockNum> <outputDir>",
		Category:  "MISCELLANEOUS COMMANDS",
		Description: `
The makecache command generates an ethash cache in <outputDir>.

This command exists to support the system testing project.
Regular users do not need to execute it.
`,
	}
	makedagCommand = cli.Command{
		Action:    utils.MigrateFlags(makedag),
		Name:      "makedag",
		Usage:     "Generate ethash mining DAG (for testing)",
		ArgsUsage: "<blockNum> <outputDir>",
		Category:  "MISCELLANEOUS COMMANDS",
		Description: `
The makedag command generates an ethash DAG in <outputDir>.

This command exists to support the system testing project.
Regular users do not need to execute it.
`,
	}
	versionCommand = cli.Command{
		Action:    utils.MigrateFlags(version),
		Name:      "version",
		Usage:     "Print version numbers",
		ArgsUsage: " ",
		Category:  "MISCELLANEOUS COMMANDS",
		Description: `
The output of this command is supposed to be machine-readable.
`,
	}
	licenseCommand = cli.Command{
		Action:    utils.MigrateFlags(license),
		Name:      "license",
		Usage:     "Display license information",
		ArgsUsage: " ",
		Category:  "MISCELLANEOUS COMMANDS",
	}

	// append by Albert·Gou
	createInitDKSCommand = cli.Command{
		Action:    utils.MigrateFlags(createInitDKS),
		Name:      "initdks",
		Usage:     "Generate the initial distributed key share",
		ArgsUsage: "",
		Category:  "MEDIATOR COMMANDS",
		Description: `
The output of this command will be used to initialize the DistKeyGenerator.
`,
	}

	// append by Albert·Gou
	nodeInfoCommand = cli.Command{
		Action:    utils.MigrateFlags(getNodeInfo),
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
)

// makecache generates an ethash verification cache into the provided folder.
func makecache(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 2 {
		utils.Fatalf(`Usage: gptn makecache <block number> <outputdir>`)
	}
	_, err := strconv.ParseUint(args[0], 0, 64)
	if err != nil {
		utils.Fatalf("Invalid block number: %v", err)
	}

	return nil
}

// makedag generates an ethash mining DAG into the provided folder.
func makedag(ctx *cli.Context) error {
	args := ctx.Args()
	if len(args) != 2 {
		utils.Fatalf(`Usage: gptn makedag <block number> <outputdir>`)
	}
	_, err := strconv.ParseUint(args[0], 0, 64)
	if err != nil {
		utils.Fatalf("Invalid block number: %v", err)
	}

	return nil
}

func version(ctx *cli.Context) error {
	fmt.Println(strings.Title(clientIdentifier))
	fmt.Println("Version:", configure.Version)
	if gitCommit != "" {
		fmt.Println("Git Commit:", gitCommit)
	}
	fmt.Println("Architecture:", runtime.GOARCH)
	fmt.Println("Protocol Versions:", ptn.ProtocolVersions)
	fmt.Println("Network Id:", ptn.DefaultConfig.NetworkId)
	fmt.Println("Go Version:", runtime.Version())
	fmt.Println("Operating System:", runtime.GOOS)
	fmt.Printf("GOPATH=%s\n", os.Getenv("GOPATH"))
	fmt.Printf("GOROOT=%s\n", runtime.GOROOT())
	return nil
}

func license(_ *cli.Context) error {
	fmt.Println(`Gptn is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Gptn is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with gptn. If not, see <http://www.gnu.org/licenses/>.`)
	return nil
}

// author Albert·Gou
func createInitDKS(ctx *cli.Context) error {
	suite := bn256.NewSuiteG2()
	sec, pub := mp.GenInitPair(suite)

	secStr := core.ScalarToStr(sec)
	pubStr := core.PointToStr(pub)

	fmt.Println("Generate a initial distributed key share:")
	fmt.Println("{")
	fmt.Println("\tprivate key: ", secStr)
	fmt.Println("\tpublic key: ", pubStr)
	fmt.Println("}")

	return nil
}

// author Albert·Gou
func getNodeInfo(ctx *cli.Context) error {
	stack := makeFullNode(ctx)
	privateKey := stack.Config().NodeKey()
	listenAddr := stack.ListenAddr()

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	realaddr := listener.Addr().(*net.TCPAddr)

	node := discover.NewNode(
		discover.PubkeyID(&privateKey.PublicKey),
		realaddr.IP,
		uint16(realaddr.Port),
		uint16(realaddr.Port))

	fmt.Println(node.String())

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
