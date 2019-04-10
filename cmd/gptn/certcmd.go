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


package main

import (
	"fmt"
	"github.com/palletone/digital-identity/client"
	"github.com/palletone/go-palletone/cmd/utils"
	"gopkg.in/urfave/cli.v1"
)

var (
	certCommand = cli.Command{
		Name:      "cert",
		Usage:     "Manage certificate",
		ArgsUsage: "",
		Category:  "MEDIATOR COMMANDS",
		Description: `
    Palletone digital certificate management, providing for users to issue certificates, revoke certificates, query certificates and other operations
`,
		Subcommands: []cli.Command{
			//注册admin证书
			{
				Action:    utils.MigrateFlags(enrollAdmin),
				Name:      "admin",
				Usage:     "Registration administrator certificate",
				ArgsUsage: "",
				Category:  "CERT COMMANDS",
				Description: `
Send the request for the registration administrator certificate to the fabric ca server.
`,
			},
			{
				Action:    utils.MigrateFlags(enrollUser),
				Name:      "new",
				Usage:     "Registered user.",
				ArgsUsage: "",
				Category:  "CERT COMMANDS",
				Description: `
Send the registered user request to the fabric ca server.
`,
			},
			{
				Action:    utils.MigrateFlags(revoke),
				Name:      "revoke",
				Usage:     "Revoke a certificate of an address.",
				ArgsUsage: "<address>",
				Category:  "CERT COMMANDS",
				Description: `
gptn cert revoke <address>

Palletone sends a request to the fabric ca server to cancel the certificate, and CRL files are generated in the MSP directory.
`,
			},
			{
				Action:    utils.MigrateFlags(getIndentity),
				Name:      "getindentity",
				Usage:     "get a certificate indentity",
				ArgsUsage: "<address> <caname>",
				Category:  "CERT COMMANDS",
				Description: `
gptn cert getindentity <address> <caname>
Gets the certificate identity attribute based on the address and caname.
`,
			},
			{
				Action:    utils.MigrateFlags(getIndentities),
				Name:      "getindenties",
				Usage:     "get certificate indentities",
				ArgsUsage: "",
				Category:  "CERT COMMANDS",
				Description: `
gptn cert getindenties
Gets the certificate identities .
`,
			},
			{
				Action:    utils.MigrateFlags(getCaCertificateChain),
				Name:      "getcertchain",
				Usage:     "get certificate chain",
				ArgsUsage: "<caname>",
				Category:  "CERT COMMANDS",
				Description: `
gptn cert getcacertificatechain
Gets the certificate chain attribute based on the  caname.
`,
			},
		},
	}
)

func newCaGenInfo() *client.CaGenInfo {
	cainfo := client.NewCaGenInfo("14", "zk", "Hi palletOne", true, "user", "gptn.mediator1",)
	return cainfo
}


func enrollAdmin(ctx *cli.Context) error {
	cainfo := newCaGenInfo()
	err := cainfo.EnrollAdmin()
	if err != nil {
		return err
	}
	fmt.Println("Registration administrator certificate ok")

	return nil
}


func enrollUser(ctx *cli.Context) error {
	cainfo := newCaGenInfo()
	err := cainfo.Enrolluser()
	if err != nil {
		return err
	}

	return nil
}

func revoke(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		fmt.Println("No certficate to revoke")
	}

	for _,addr := range ctx.Args() {
		fmt.Println(addr)
		cainfo := newCaGenInfo()
		reson := "Forced to compromise"
		err := cainfo.Revoke(addr,reson)
		if err != nil {
			return err
		}
	}


	return nil
}

func getIndentity(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		fmt.Println("No certficate to getIndentity")
		return nil
	}
	if len(ctx.Args()) != 2 {
		fmt.Println("Please enter full address and caname")
		return nil
	}
	address := ctx.Args().First()
	caname := ctx.Args()[1]
	cainfo := newCaGenInfo()

    idtRep := cainfo.GetIndentity(address,caname)
    fmt.Println(idtRep)
	return nil
}

func getIndentities(ctx *cli.Context) error {
	cainfo := newCaGenInfo()

	idtReps := cainfo.GetIndentities()
	fmt.Println(idtReps)
	return nil
}

func getCaCertificateChain(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		fmt.Println("No certficatechain to get")
		return nil
	}
	caname := ctx.Args().First()
	cainfo := newCaGenInfo()

	idtReps,err := cainfo.GetCaCertificateChain(caname)
	if err != nil {
		return err
	}
	fmt.Println(idtReps)
	return nil
}