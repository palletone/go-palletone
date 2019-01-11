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

package flogging

/*
import (
	"github.com/palletone/go-palletone/common/log"
)

const GRPCModuleID = "grpc"

func initgrpclog() {
	//glog := MustGetLogger(GRPCModuleID)
	grpclog.SetLogger(&grpclog{glog})
}

// grpclog implements the standard Go logging interface and wraps the
// log provided by the flogging package.  This is required in order to
// replace the default log used by the grpclog package.
type grpclog struct {
	log log.ILogger
}

func (g *grpclog) Fatal(args ...interface{}) {
	g.log.Error(GRPCModuleID, args...)
}

func (g *grpclog) Fatalf(format string, args ...interface{}) {
	g.log.Errorf(format, args...)
}

func (g *grpclog) Fatalln(args ...interface{}) {
	g.log.Error(GRPCModuleID, args...)
}

// NOTE: grpclog does not support leveled logs so for now use DEBUG
func (g *grpclog) Print(args ...interface{}) {
	g.log.Debug(GRPCModuleID, args...)
}

func (g *grpclog) Printf(format string, args ...interface{}) {
	g.log.Debugf(format, args...)
}

func (g *grpclog) Println(args ...interface{}) {
	g.log.Debug(GRPCModuleID, args...)
}
*/
