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
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/palletone/go-palletone/cmd/utils"
	"github.com/palletone/go-palletone/configure"
	"github.com/palletone/go-palletone/ptn"
	"gopkg.in/urfave/cli.v1"
	"github.com/dedis/kyber/pairing/bn256"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/common/log"
	"encoding/base64"
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
		ArgsUsage: " ",
		Category:  "MEDIATOR COMMANDS",
		Description: `
The output of this command will be used to initialize a DistKeyGenerator.
`,
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
	var err error
	suite := bn256.NewSuiteG2()
	sec, pub := mp.GenInitPair(suite)

	secB, err := sec.MarshalBinary()
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}
	pubB, err := pub.MarshalBinary()
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}

	secStr := base64.RawURLEncoding.EncodeToString(secB)
	pubStr := base64.RawURLEncoding.EncodeToString(pubB)
	fmt.Println("Generate a initial distributed key share:")
	fmt.Println("{")
	fmt.Println("\tprivate key: ", secStr)
	fmt.Println("\tpublic key: ", pubStr)
	fmt.Println("}")

	return nil
}
