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
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package main

import (
	"io/ioutil"
	"log"
	"golang.org/x/crypto/ed25519"
)

func main() {
	_, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Panic(err)
	}
	err = ioutil.WriteFile("vrf.vrfsecret", sk, 0600)
	if err != nil {
		log.Panic(err)
	}
}
