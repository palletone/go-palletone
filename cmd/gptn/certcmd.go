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

			// 列出当前区块链所有mediator的地址
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
		},
	}
)

func newCaGenInfo() *client.CaGenInfo {
	cainfo := client.NewCaGenInfo("11", "zk", "Hi palletOne", true, "user", "gptn.mediator1",)
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

