/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */
package config

import (
"fmt"
"io/ioutil"
"os"

"gopkg.in/yaml.v2"
)

type CAConfig struct {
	CryptoConfig             `yaml:"crypto"`
	URL               string `yaml:"url"`
	SkipTLSValidation bool   `yaml:"skipTLSValidation"`
	MspID             string `yaml:"mspId"`
	FilePath          string `yaml:"filepath"`
	Admin             string `yaml:"admin"`
	Adminpw           string `yaml:"adminpw"`
}

type CryptoConfig struct {
	Family    string `yaml:"family"`
	Algorithm string `yaml:"algorithm"`
	Hash      string `yaml:"hash"`
}

func NewCAConfig(path string) (*CAConfig, error) {

	data, err := ioutil.ReadFile(path)

	if err != nil {
		return nil, err
	}
	config := new(CAConfig)
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func IsPathExists(filepath string) (bool, error) {
	fi, err := os.Stat(filepath)
	if err != nil {
		return false, err
	}
	if !fi.IsDir() {
		return false, fmt.Errorf("THE PATH: %s is not directory", filepath)
	}
	return true, nil
}
