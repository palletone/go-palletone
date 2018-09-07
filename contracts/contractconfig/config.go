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

package contractconfig

var DefaultConfig = Config{
	Executetimeout:30,
	Mode:"net",
}

type id struct{
	path   string
	name   string
}

type Config struct {
	Id       id
	/*
	# Timeout duration for Invoke and Init calls to prevent runaway.
	# This timeout is used by all chaincodes in all the channels, including
	# system chaincodes.
	# Note that during Invoke, if the image is not available (e.g. being
	# cleaned up when in development environment), the peer will automatically
	# build the image, which might take more time. In production environment,
	# the chaincode image is unlikely to be deleted, so the timeout could be
	# reduced accordingly.
	*/
	Executetimeout int
	Mode string
	builder  string
	pull     bool
}

