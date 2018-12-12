// Copyright 2016 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"crypto/rand"
	"math/big"
	"testing"
)

const (
	ipcAPIs  = "admin:1.0 debug:1.0 gptn:1.0 miner:1.0 net:1.0 personal:1.0 rpc:1.0 shh:1.0 txpool:1.0 web3:1.0"
	httpAPIs = "gptn:1.0 net:1.0 rpc:1.0 web3:1.0"
)

// Tests that a node embedded within a console can be started up properly and
// then terminated by closing the input stream.
func TestConsoleWelcome(t *testing.T) {
	// Start a gptn console, make sure it's cleaned up and terminate the console
	gptn := runGptn(t, "console")
	// Gather all the infos the welcome message needs to contain

	gptn.Expect("last")
	//gptn.ExpectExitConsoleWelcome()
}

func TestPeerToPeer(t *testing.T) {
	//第一个节点的 pnode://6632b753a9e83bfad296f31689e7c7566ba921babc19ff8fdb99a617f494b0afc84e06b678bb567600dda2fd78cc69aaa9424051efc95c51961ffecbb30aace5

	//第二个节点添加第一个节点的 pnode
	//_ = runGptn(t,"--exec","admin.addPeer(\"pnode://6632b753a9e83bfad296f31689e7c7566ba921babc19ff8fdb99a617f494b0afc84e06b678bb567600dda2fd78cc69aaa9424051efc95c51961ffecbb30aace5@[::]:30303\")",
	//	"attach","\\\\.\\pipe\\gptn.ipc")
	//第一个节点查看 addmin.peers
	//gptn:= runGptn(t,"--exec","admin.peers[0].id","attach","\\\\.\\pipe\\gptn1.ipc")
	//gptn.Expect("\"ffe883a450a8a6fc3316113b62cbd6fa06f7947bf7b4285a85b287b0bd79b2f512106ffcf20bbdb3e012890c703b6d5fefdb06332e24c9894ccb0d4ea667d1e2\"")
}

/*
// Tests that a console can be attached to a running node via various means.
func TestIPCAttachWelcome(t *testing.T) {
	// Configure the instance for IPC attachement
	//coinbase := "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
	var ipc string
	if runtime.GOOS == "windows" {
		ipc = `\\.\pipe\gptn` + strconv.Itoa(trulyRandInt(100000, 999999))
	} else {
		//ws := "./data1/"//tmpdir(t)
		//defer os.RemoveAll(ws)
		ipc = "./data1/gptn.ipc"//filepath.Join(ws, "gptn.ipc")
	}
	// Note: we need --shh because testAttachWelcome checks for default
	// list of ipc modules and shh is included there.
	gptn := runGptn(t,"--maxpeers", "0", "--nodiscover", "--nat", "none")

	time.Sleep(10 * time.Second) // Simple way to wait for the RPC endpoint to open
	testAttachWelcome(t, gptn, "ipc:"+ipc, ipcAPIs)

	gptn.Interrupt()
	//gptn.Expect("Welcome to the Gpan JavaScript console")
	gptn.ExpectExitIPCAttachWelcome()
}

func TestHTTPAttachWelcome(t *testing.T) {
	//coinbase := "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
	port := strconv.Itoa(trulyRandInt(1024, 65536)) // Yeah, sometimes this will fail, sorry :P
	//gptn := runGptn(t,
	//	"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
	//	"--etherbase", coinbase, "--rpc", "--rpcport", port)
	gptn := runGptn(t)

	time.Sleep(2 * time.Second) // Simple way to wait for the RPC endpoint to open
	testAttachWelcome(t, gptn, "http://localhost:"+port, httpAPIs)

	gptn.Interrupt()
	gptn.ExpectExit()
}

func TestWSAttachWelcome(t *testing.T) {
	coinbase := "0x8605cdbbdb6d264aa742e77020dcbc58fcdce182"
	port := strconv.Itoa(trulyRandInt(1024, 65536)) // Yeah, sometimes this will fail, sorry :P

	gptn := runGptn(t,
		"--port", "0", "--maxpeers", "0", "--nodiscover", "--nat", "none",
		"--etherbase", coinbase, "--ws", "--wsport", port)

	time.Sleep(2 * time.Second) // Simple way to wait for the RPC endpoint to open
	testAttachWelcome(t, gptn, "ws://localhost:"+port, httpAPIs)

	gptn.Interrupt()
	gptn.ExpectExit()
}
*/
func testAttachWelcome(t *testing.T, gptn *testgptn, endpoint, apis string) {
	// Attach to a running gptn note and terminate immediately
	attach := runGptn(t, "attach", endpoint)
	defer attach.ExpectExit()
	attach.CloseStdin()

	// Gather all the infos the welcome message needs to contain
	//attach.SetTemplateFunc("goos", func() string { return runtime.GOOS })
	//attach.SetTemplateFunc("goarch", func() string { return runtime.GOARCH })
	//attach.SetTemplateFunc("gover", runtime.Version)
	////attach.SetTemplateFunc("gethver", func() string { return configure.Version })
	//attach.SetTemplateFunc("etherbase", func() string { return gptn.Etherbase })
	//attach.SetTemplateFunc("niltime", func() string { return time.Unix(0, 0).Format(time.RFC1123) })
	//attach.SetTemplateFunc("ipc", func() bool { return strings.HasPrefix(endpoint, "ipc") })
	//attach.SetTemplateFunc("datadir", func() string { return gptn.Datadir })
	//attach.SetTemplateFunc("apis", func() string { return apis })

	// Verify the actual welcome message to the required template
	//	attach.Expect(`
	//Welcome to the Geth JavaScript console!
	//
	//instance: Geth/v{{gethver}}/{{goos}}-{{goarch}}/{{gover}}
	//coinbase: {{etherbase}}
	//at block: 0 ({{niltime}}){{if ipc}}
	// datadir: {{datadir}}{{end}}
	// modules: {{apis}}
	//
	//> {{.InputLine "exit" }}
	//`)
	//	attach.ExpectExit()
}

// trulyRandInt generates a crypto random integer used by the console tests to
// not clash network ports with other tests running cocurrently.
func trulyRandInt(lo, hi int) int {
	num, _ := rand.Int(rand.Reader, big.NewInt(int64(hi-lo)))
	return int(num.Int64()) + lo
}
