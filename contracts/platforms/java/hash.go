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
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/palletone/go-palletone/common/log"
	ccutil "github.com/palletone/go-palletone/contracts/platforms/util"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	"github.com/palletone/go-palletone/core/vmContractPub/util"
)

//var log = flogging.MustGetLogger("java/hash")

//collectChaincodeFiles collects chaincode files and generates hashcode for the
//package.
//NOTE: for dev mode, user builds and runs chaincode manually. The name provided
//by the user is equivalent to the path. This method will treat the name
//as codebytes and compute the hash from it. ie, user cannot run the chaincode
//with the same (name, input, args)
func collectChaincodeFiles(spec *pb.ChaincodeSpec, tw *tar.Writer) (string, error) {
	if spec == nil {
		return "", errors.New("Cannot collect chaincode files from nil spec")
	}

	chaincodeID := spec.ChaincodeId
	if chaincodeID == nil || chaincodeID.Path == "" {
		return "", errors.New("Cannot collect chaincode files from empty chaincode path")
	}

	codepath := chaincodeID.Path

	var err error
	if !strings.HasPrefix(codepath, "/") {
		wd := ""
		wd, err = os.Getwd()
		codepath = wd + "/" + codepath
	}

	if err != nil {
		return "", fmt.Errorf("Error getting code %s", err)
	}

	if err = ccutil.IsCodeExist(codepath); err != nil {
		return "", fmt.Errorf("code does not exist %s", err)
	}

	var hash []byte

	//install will not have inputs and we don't have to collect hash for it
	if spec.Input == nil || len(spec.Input.Args) == 0 {
		log.Debugf("not using input for hash computation for %v ", chaincodeID)
	} else {
		inputbytes, err2 := proto.Marshal(spec.Input)
		if err2 != nil {
			return "", fmt.Errorf("error marshaling constructor: %s", err)
		}
		hash = util.GenerateHashFromSignature(codepath, inputbytes)
	}

	hash, err = ccutil.HashFilesInDir("", codepath, hash, tw)
	if err != nil {
		return "", fmt.Errorf("could not get hashcode for %s - %s", codepath, err)
	}

	return hex.EncodeToString(hash[:]), nil
}
