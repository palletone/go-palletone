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
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/palletone/go-palletone/core/vmContractPub/metadata"
)

func TestUtil_DockerfileTemplateParser(t *testing.T) {
	expected := "FROM foo:" + getArch() + "-" + metadata.Version
	actual := ParseDockerfileTemplate("FROM foo:$(ARCH)-$(PROJECT_VERSION)")
	assert.Equal(t, expected, actual, "Error parsing Dockerfile Template. Expected \"%s\", got \"%s\"",
		expected, actual)
}

func TestUtil_GetDockerfileFromConfig(t *testing.T) {
	expected := "FROM " + metadata.DockerNamespace + ":" + getArch() + "-" + metadata.Version
	path := "dt"
	viper.Set(path, "FROM $(DOCKER_NS):$(ARCH)-$(PROJECT_VERSION)")
	actual := GetDockerfileFromConfig(path)
	assert.Equal(t, expected, actual, "Error parsing Dockerfile Template. Expected \"%s\", got \"%s\"",
		expected, actual)
}

func TestUtil_GetDockertClient(t *testing.T) {
	viper.Set("vm.endpoint", "unix:///var/run/docker.sock")
	_, err := NewDockerClient()
	assert.NoError(t, err, "Error getting docker client")
}
