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

package log

//对于有些情况，我们只希望在开发环境的时候记录Log，而在生产环境不做任何记录
type NothingLogger struct{}

func (n *NothingLogger) Trace(msg string, ctx ...interface{})  {}
func (n *NothingLogger) Debug(msg string, ctx ...interface{})  {}
func (n *NothingLogger) Debugf(msg string, ctx ...interface{}) {}
func (n *NothingLogger) Info(msg string, ctx ...interface{})   {}
func (n *NothingLogger) Infof(msg string, ctx ...interface{})  {}
func (n *NothingLogger) Warn(msg string, ctx ...interface{})   {}
func (n *NothingLogger) Warnf(msg string, ctx ...interface{})  {}
func (n *NothingLogger) Error(msg string, ctx ...interface{})  {}
func (n *NothingLogger) Errorf(msg string, ctx ...interface{}) {}
func (n *NothingLogger) Crit(msg string, ctx ...interface{})   {}
