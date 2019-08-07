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

package controller

import (
	"fmt"
	"io"
	"sync"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/vm/api"
	"github.com/palletone/go-palletone/vm/ccintf"
	"github.com/palletone/go-palletone/vm/dockercontroller"
	"github.com/palletone/go-palletone/vm/inproccontroller"
	"golang.org/x/net/context"
)

type refCountedLock struct {
	refCount int
	lock     *sync.RWMutex
}

//VMController - manages VMs
//   . abstract construction of different types of VMs (we only care about Docker for now)
//   . manage lifecycle of VM (start with build, start, stop ...
//     eventually probably need fine grained management)
type VMController struct {
	sync.RWMutex
	// Handlers for each chaincode
	containerLocks map[string]*refCountedLock
}

//singleton...acess through NewVMController
var vmcontroller *VMController

//constants for supported containers
const (
	DOCKER = "Docker"
	SYSTEM = "System"
)

//NewVMController - creates/returns singleton
func init() {
	vmcontroller = new(VMController)
	vmcontroller.containerLocks = make(map[string]*refCountedLock)
}

func (vmc *VMController) newVM(typ string) api.VM {
	var (
		v api.VM
	)

	switch typ {
	case DOCKER:
		v = dockercontroller.NewDockerVM()
	case SYSTEM:
		v = &inproccontroller.InprocVM{}
	default:
		v = &dockercontroller.DockerVM{}
	}
	return v
}

func (vmc *VMController) lockContainer(id string) {
	//get the container lock under global lock
	vmcontroller.Lock()
	var refLck *refCountedLock
	var ok bool
	if refLck, ok = vmcontroller.containerLocks[id]; !ok {
		refLck = &refCountedLock{refCount: 1, lock: &sync.RWMutex{}}
		vmcontroller.containerLocks[id] = refLck
	} else {
		refLck.refCount++
		log.Debugf("refcount %d (%s)", refLck.refCount, id)
	}
	vmcontroller.Unlock()
	log.Debugf("waiting for container(%s) lock", id)
	refLck.lock.Lock()
	log.Debugf("got container (%s) lock", id)
}

func (vmc *VMController) unlockContainer(id string) {
	vmcontroller.Lock()
	if refLck, ok := vmcontroller.containerLocks[id]; ok {
		if refLck.refCount <= 0 {
			panic("refcnt <= 0")
		}
		refLck.lock.Unlock()
		if refLck.refCount--; refLck.refCount == 0 {
			log.Debugf("container lock deleted(%s)", id)
			delete(vmcontroller.containerLocks, id)
		}
	} else {
		log.Debugf("no lock to unlock(%s)!!", id)
	}
	vmcontroller.Unlock()
}

//VMCReqIntf - all requests should implement this interface.
//The context should be passed and tested at each layer till we stop
//note that we'd stop on the first method on the stack that does not
//take context
type VMCReqIntf interface {
	do(ctxt context.Context, v api.VM) VMCResp
	getCCID() ccintf.CCID
}

//VMCResp - response from requests. resp field is a anon interface.
//It can hold any response. err should be tested first
type VMCResp struct {
	Err  error
	Resp interface{}
}

//CreateImageReq - properties for creating an container image
type CreateImageReq struct {
	ccintf.CCID
	Reader io.Reader
	Args   []string
	Env    []string
}

func (bp CreateImageReq) do(ctxt context.Context, v api.VM) VMCResp {
	var resp VMCResp

	if err := v.Deploy(ctxt, bp.CCID, bp.Args, bp.Env, bp.Reader); err != nil {
		resp = VMCResp{Err: err}
	} else {
		resp = VMCResp{}
	}

	return resp
}

func (bp CreateImageReq) getCCID() ccintf.CCID {
	return bp.CCID
}

//StartImageReq - properties for starting a container.
type StartImageReq struct {
	ccintf.CCID
	Builder       api.BuildSpecFactory
	Args          []string
	Env           []string
	FilesToUpload map[string][]byte
	PrelaunchFunc api.PrelaunchFunc
}

func (si StartImageReq) do(ctxt context.Context, v api.VM) VMCResp {
	var resp VMCResp

	if err := v.Start(ctxt, si.CCID, si.Args, si.Env, si.FilesToUpload, si.Builder, si.PrelaunchFunc); err != nil {
		resp = VMCResp{Err: err}
	} else {
		resp = VMCResp{}
	}

	return resp
}

func (si StartImageReq) getCCID() ccintf.CCID {
	return si.CCID
}

//StopImageReq - properties for stopping a container.
type StopImageReq struct {
	ccintf.CCID
	Timeout uint
	//by default we will kill the container after stopping
	Dontkill bool
	//by default we will remove the container after killing
	Dontremove bool
}

func (si StopImageReq) do(ctxt context.Context, v api.VM) VMCResp {
	var resp VMCResp

	if err := v.Stop(ctxt, si.CCID, si.Timeout, si.Dontkill, si.Dontremove); err != nil {
		resp = VMCResp{Err: err}
	} else {
		resp = VMCResp{}
	}

	return resp
}

func (si StopImageReq) getCCID() ccintf.CCID {
	return si.CCID
}

//DestroyImageReq - properties for stopping a container.
type DestroyImageReq struct {
	ccintf.CCID
	Timeout uint
	Force   bool
	NoPrune bool
}

func (di DestroyImageReq) do(ctxt context.Context, v api.VM) VMCResp {
	var resp VMCResp

	if err := v.Destroy(ctxt, di.CCID, di.Force, di.NoPrune); err != nil {
		resp = VMCResp{Err: err}
	} else {
		resp = VMCResp{}
	}

	return resp
}

func (di DestroyImageReq) getCCID() ccintf.CCID {
	return di.CCID
}

//VMCProcess should be used as follows
//   . construct a context
//   . construct req of the right type (e.g., CreateImageReq)
//   . call it in a go routine
//   . process response in the go routing
//context can be canceled. VMCProcess will try to cancel calling functions if it can
//For instance docker clients api's such as BuildImage are not cancelable.
//In all cases VMCProcess will wait for the called go routine to return
func VMCProcess(ctxt context.Context, vmtype string, req VMCReqIntf) (interface{}, error) {
	v := vmcontroller.newVM(vmtype)

	if v == nil {
		return nil, fmt.Errorf("Unknown VM type %s", vmtype)
	}

	c := make(chan struct{})
	var resp interface{}
	go func() {
		defer close(c)

		id, err := v.GetVMName(req.getCCID(), nil)
		if err != nil {
			resp = VMCResp{Err: err}
			return
		}
		vmcontroller.lockContainer(id)
		resp = req.do(ctxt, v)
		vmcontroller.unlockContainer(id)
	}()

	select {
	case <-c:
		return resp, nil
	case <-ctxt.Done():
		//TODO cancel req.do ... (needed) ?
		<-c
		return nil, ctxt.Err()
	}
}
