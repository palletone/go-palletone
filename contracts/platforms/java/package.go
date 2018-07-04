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

package java

import (
	"archive/tar"
	"fmt"
	"strings"
	"errors"

	cutil "github.com/palletone/go-palletone/vm/common"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
)

//tw is expected to have the chaincode in it from GenerateHashcode.
//This method will just package the dockerfile
func writeChaincodePackage(spec *pb.ChaincodeSpec, tw *tar.Writer) error {
	urlLocation := spec.ChaincodeId.Path
	if urlLocation == "" {
		return errors.New("ChaincodeSpec's path/URL cannot be empty")
	}

	if strings.LastIndex(urlLocation, "/") == len(urlLocation)-1 {
		urlLocation = urlLocation[:len(urlLocation)-1]
	}

	err := cutil.WriteJavaProjectToPackage(tw, urlLocation)
	if err != nil {
		return fmt.Errorf("Error writing Chaincode package contents: %s", err)
	}

	return nil
}
