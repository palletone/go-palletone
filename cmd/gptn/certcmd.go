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
			//注册用户证书，默认type:user  ECert:true
			{
				Action:    utils.MigrateFlags(enrollUser),
				Name:      "new",
				Usage:     "Registered user <address><name><data><affiliation>",
				ArgsUsage: "<address><name><data><affiliation>",
				Category:  "CERT COMMANDS",
				Description: `
Send the registered user request to the fabric ca server.
`,
			},
			{
				Action:    utils.MigrateFlags(revoke),
				Name:      "revoke",
				Usage:     "Revoke a certificate of an address <address><reason>",
				ArgsUsage: "<address><reason>",
				Category:  "CERT COMMANDS",
				Description: `
gptn cert revoke <address><reason>

Palletone sends a request to the fabric ca server to cancel the certificate, and CRL files are generated in the MSP directory.
`,
			},
			{
				Action:    utils.MigrateFlags(getIndentity),
				Name:      "getindentity",
				Usage:     "get a certificate indentity <address> <caname>",
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
				Usage:     "get certificate chain <caname>",
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

//func newCaGenInfo() *client.CaGenInfo {
//	cainfo := client.NewCaGenInfo("15", "zk", "Hi palletOne", true, "user", "gptn.mediator1",)
//	return cainfo
//}

func enrollAdmin(ctx *cli.Context) error {
	cainfo := client.CaGenInfo{}
	err := cainfo.EnrollAdmin()
	if err != nil {
		return err
	}
	fmt.Println("Registration administrator certificate ok")

	return nil
}

func enrollUser(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		fmt.Println("Please enter parameters <address><name><data><affiliation>")
		return nil
	}

	if len(ctx.Args()) != 4 {
		fmt.Println("Registered user certificate should fill in the <address><name><data><affiliation> ")
		return nil
	}

	address := ctx.Args().First()
	name := ctx.Args()[1]
	data := ctx.Args()[2]
	affiliation := ctx.Args()[3]
	ty := "user"

	certinfo := NewCertInfo(address, name, data, ty, affiliation, true)
	err := GenCert(*certinfo)
	if err != nil {
		return err
	}
	fmt.Println(address + "  Registered  certificate  OK")
	return nil
}

func revoke(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		fmt.Println("No certficate to revoke")
		return nil
	}

	if len(ctx.Args()) == 1 {
		address := ctx.Args().First()
		reason := "Forced to compromise"
		err := RevokeCert(address,reason)
		if err != nil {
			return err
		}
		fmt.Println(address + "  Revoked  certificate  OK ,Reason is Forced to compromise.")
		return nil
	}
	address := ctx.Args().First()
	reason := ctx.Args()[1]

	err := RevokeCert(address, reason)
	fmt.Println(address + "  Revoked  certificate  OK")
	if err != nil {
		return err
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

	idtRep, err := GetIndentity(address, caname)
	if err != nil {
		return err
	}
	fmt.Println(idtRep)
	return nil
}

func getIndentities(ctx *cli.Context) error {
	idtReps,err := GetIndentities()
	if err != nil {
		return err
	}
	fmt.Println(idtReps)
	return nil
}

func getCaCertificateChain(ctx *cli.Context) error {
	if len(ctx.Args()) == 0 {
		fmt.Println("No certficatechain to get,Please enter parameters <caname>.")
		return nil
	}
	caname := ctx.Args().First()

	certchain, err := GetCaCertificateChain(caname)
	if err != nil {
		return err
	}
	fmt.Println(certchain)
	return nil
}
