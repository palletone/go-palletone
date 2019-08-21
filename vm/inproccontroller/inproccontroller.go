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

package inproccontroller

import (
	"fmt"
	"io"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/contracts/shim"
	pb "github.com/palletone/go-palletone/core/vmContractPub/protos/peer"
	container "github.com/palletone/go-palletone/vm/api"
	"github.com/palletone/go-palletone/vm/ccintf"
	"golang.org/x/net/context"
)

type inprocContainer struct {
	chaincode shim.Chaincode
	running   bool
	args      []string
	env       []string
	stopChan  chan struct{}
}

var (
	//log        = flogging.MustGetLogger("inproccontroller")
	typeRegistry     = make(map[string]*inprocContainer)
	instRegistry     = make(map[string]*inprocContainer)
	_shimStartInProc = shim.StartInProc
	_logErrorf       = log.Errorf
)

// errors

//SysCCRegisteredErr registered error
type SysCCRegisteredErr string

func (s SysCCRegisteredErr) Error() string {
	return fmt.Sprintf("%s already registered", string(s))
}

//Register registers system chaincode with given path. The deploy should be called to initialize
func Register(path string, cc shim.Chaincode) error {
	tmp := typeRegistry[path]
	if tmp != nil {
		return SysCCRegisteredErr(path)
	}

	typeRegistry[path] = &inprocContainer{chaincode: cc}
	return nil
}

//InprocVM is a vm. It is identified by a executable name
type InprocVM struct {
	//id string
}

func (vm *InprocVM) getInstance(_ context.Context, ipctemplate *inprocContainer, instName string, args []string, env []string) (*inprocContainer, error) {
	ipc := instRegistry[instName]
	if ipc != nil {
		log.Warnf("chaincode instance exists for %s", instName)
		return ipc, fmt.Errorf("chaincode instance exists for %s", instName)
	}
	ipc = &inprocContainer{args: args, env: env, chaincode: ipctemplate.chaincode, stopChan: make(chan struct{})}
	instRegistry[instName] = ipc
	log.Debugf("chaincode instance created for %s", instName)
	return ipc, nil
}

//Deploy verifies chaincode is registered and creates an instance for it. Currently only one instance can be created
func (vm *InprocVM) Deploy(ctxt context.Context, ccid ccintf.CCID, args []string, env []string, reader io.Reader) error {
	path := ccid.ChaincodeSpec.ChaincodeId.Path

	ipctemplate := typeRegistry[path]
	if ipctemplate == nil {
		return fmt.Errorf(fmt.Sprintf("%s not registered. Please register the system chaincode in inprocinstances.go", path))
	}

	if ipctemplate.chaincode == nil {
		return fmt.Errorf(fmt.Sprintf("%s system chaincode does not contain chaincode instance", path))
	}

	instName, _ := vm.GetVMName(ccid, nil)
	_, err := vm.getInstance(ctxt, ipctemplate, instName, args, env)

	//FUTURE ... here is where we might check code for safety
	log.Debugf("registered : %s", path)

	return err
}

func (ipc *inprocContainer) launchInProc(ctxt context.Context, id string, args []string, env []string, ccSupport ccintf.CCSupport) error {
	peerRcvCCSend := make(chan *pb.ChaincodeMessage)
	ccRcvPeerSend := make(chan *pb.ChaincodeMessage)
	var err error
	ccchan := make(chan struct{}, 1)
	ccsupportchan := make(chan struct{}, 1)
	go func() {
		defer close(ccchan)
		log.Debugf("chaincode started for %s", id)
		if args == nil {
			args = ipc.args
		}
		if env == nil {
			env = ipc.env
		}
		err = _shimStartInProc(env, args, ipc.chaincode, ccRcvPeerSend, peerRcvCCSend)
		if err != nil {
			err = fmt.Errorf("chaincode-support ended with err: %s", err)
			_logErrorf("%s", err)
		} else {
			log.Debugf("chaincode ended with for  %s", id)
		}
	}()

	go func() {
		defer close(ccsupportchan)
		inprocStream := newInProcStream(peerRcvCCSend, ccRcvPeerSend)
		log.Debugf("chaincode-support started for  %s", id)
		err = ccSupport.HandleChaincodeStream(ctxt, inprocStream)
		if err != nil {
			err = fmt.Errorf("chaincode ended with err: %s", err)
			_logErrorf("%s", err)
		} else {
			log.Debugf("chaincode-support ended with for  %s", id)
		}
	}()

	select {
	case <-ccchan:
		close(peerRcvCCSend)
		log.Debugf("chaincode %s quit", id)
	case <-ccsupportchan:
		close(ccRcvPeerSend)
		log.Debugf("chaincode support %s quit", id)
	case <-ipc.stopChan:
		close(ccRcvPeerSend)
		close(peerRcvCCSend)
		log.Debugf("chaincode %s stopped", id)
	}
	return err
}

//Start starts a previously registered system codechain
func (vm *InprocVM) Start(ctxt context.Context, ccid ccintf.CCID, args []string, env []string, filesToUpload map[string][]byte, builder container.BuildSpecFactory, prelaunchFunc container.PrelaunchFunc) error {
	path := ccid.ChaincodeSpec.ChaincodeId.Path

	ipctemplate := typeRegistry[path]

	if ipctemplate == nil {
		return fmt.Errorf(fmt.Sprintf("%s not registered", path))
	}

	instName, _ := vm.GetVMName(ccid, nil)

	ipc, err := vm.getInstance(ctxt, ipctemplate, instName, args, env)

	if err != nil {
		return fmt.Errorf(fmt.Sprintf("could not create instance for %s", instName))
	}

	if ipc.running {
		return fmt.Errorf(fmt.Sprintf("chaincode running %s", path))
	}

	//TODO VALIDITY CHECKS ?

	ccSupport, ok := ctxt.Value(ccintf.GetCCHandlerKey()).(ccintf.CCSupport)
	if !ok || ccSupport == nil {
		return fmt.Errorf("in-process communication generator not supplied")
	}

	if prelaunchFunc != nil {
		if err = prelaunchFunc(); err != nil {
			return err
		}
	}

	ipc.running = true

	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorf("caught panic from chaincode  %s", instName)
			}
		}()
		ipc.launchInProc(ctxt, instName, args, env, ccSupport)
	}()

	return nil
}

//Stop stops a system codechain
func (vm *InprocVM) Stop(ctxt context.Context, ccid ccintf.CCID, timeout uint, dontkill bool, dontremove bool) error {
	path := ccid.ChaincodeSpec.ChaincodeId.Path

	ipctemplate := typeRegistry[path]
	if ipctemplate == nil {
		return fmt.Errorf("%s not registered", path)
	}

	instName, _ := vm.GetVMName(ccid, nil)

	ipc := instRegistry[instName]

	if ipc == nil {
		return fmt.Errorf("%s not found", instName)
	}

	if !ipc.running {
		return fmt.Errorf("%s not running", instName)
	}

	ipc.stopChan <- struct{}{}

	delete(instRegistry, instName)
	//TODO stop
	return nil
}

//Destroy destroys an image
func (vm *InprocVM) Destroy(ctxt context.Context, ccid ccintf.CCID, force bool, noprune bool) error {
	//not implemented
	return nil
}

// GetVMName ignores the peer and network name as it just needs to be unique in
// process.  It accepts a format function parameter to allow different
// formatting based on the desired use of the name.
func (vm *InprocVM) GetVMName(ccid ccintf.CCID, format func(string) (string, error)) (string, error) {
	name := ccid.GetName()
	if format != nil {
		formattedName, err := format(name)
		if err != nil {
			return formattedName, err
		}
		name = formattedName
	}
	return name, nil
}
