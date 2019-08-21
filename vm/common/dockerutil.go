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
 * Copyright IBM Corp. All Rights Reserved.
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package util

import (
	"github.com/fsouza/go-dockerclient"
	"github.com/palletone/go-palletone/common/log"
	cfg "github.com/palletone/go-palletone/contracts/contractcfg"
	"github.com/palletone/go-palletone/core/vmContractPub/metadata"
	"github.com/spf13/viper"
	"runtime"
	"strings"
)

//NewDockerClient creates a docker client
func NewDockerClient() (client *docker.Client, err error) {
	//endpoint := viper.GetString("vm.endpoint")
	endpoint := cfg.GetConfig().VmEndpoint
	log.Infof("NewDockerClient enter, endpoint:%s", endpoint)

	tlsenabled := viper.GetBool("vm.docker.tls.enabled")
	if tlsenabled {
		cert := "" // config.GetPath("vm.docker.tls.cert.file")
		key := ""  //config.GetPath("vm.docker.tls.key.file")
		ca := ""   //config.GetPath("vm.docker.tls.ca.file")
		client, err = docker.NewTLSClient(endpoint, cert, key, ca)
	} else {
		client, err = docker.NewClient(endpoint)
	}
	return
}

// Our docker images retrieve $ARCH via "uname -m", which is typically "x86_64" for, well, x86_64.
// However, GOARCH uses "amd64".  We therefore need to normalize any discrepancies between "uname -m"
// and GOARCH here.
var archRemap = map[string]string{
	"amd64": "x86_64",
}

func getArch() string {
	if remap, ok := archRemap[runtime.GOARCH]; ok {
		return remap
	} else {
		return runtime.GOARCH
	}
}

func ParseDockerfileTemplate(template string) string {
	r := strings.NewReplacer(
		"$(ARCH)", getArch(),
		"$(PROJECT_VERSION)", metadata.Version,
		"$(BASE_VERSION)", metadata.BaseVersion,
		"$(DOCKER_NS)", metadata.DockerNamespace,
		"$(BASE_DOCKER_NS)", metadata.BaseDockerNamespace)

	return r.Replace(template)
}

//func GetDockerfileFromConfig(path string) string {
//	if path == "chaincode.builder" {
//		ParseDockerfileTemplate(cfg.GetConfig().ContractBuilder)
//	}
//
//	return ParseDockerfileTemplate(viper.GetString(path))
//}
