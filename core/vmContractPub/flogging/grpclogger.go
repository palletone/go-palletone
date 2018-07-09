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

import (
	"github.com/op/go-logging"
	"google.golang.org/grpc/grpclog"
)

const GRPCModuleID = "grpc"

func initgrpclogger() {
	glogger := MustGetLogger(GRPCModuleID)
	grpclog.SetLogger(&grpclogger{glogger})
}

// grpclogger implements the standard Go logging interface and wraps the
// logger provided by the flogging package.  This is required in order to
// replace the default log used by the grpclog package.
type grpclogger struct {
	logger *logging.Logger
}

func (g *grpclogger) Fatal(args ...interface{}) {
	g.logger.Fatal(args...)
}

func (g *grpclogger) Fatalf(format string, args ...interface{}) {
	g.logger.Fatalf(format, args...)
}

func (g *grpclogger) Fatalln(args ...interface{}) {
	g.logger.Fatal(args...)
}

// NOTE: grpclog does not support leveled logs so for now use DEBUG
func (g *grpclogger) Print(args ...interface{}) {
	g.logger.Debug(args...)
}

func (g *grpclogger) Printf(format string, args ...interface{}) {
	g.logger.Debugf(format, args...)
}

func (g *grpclogger) Println(args ...interface{}) {
	g.logger.Debug(args...)
}
