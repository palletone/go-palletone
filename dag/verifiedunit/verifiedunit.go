/**
@version 0.1
@author albert·gou
@time June 22, 2018
@brief 验证单元定义
*/

package verifiedunit

import "time"

type VerifiedUnit struct {
	PreVerifiedUnit *VerifiedUnit // 前一个验证单元的hash
	MediatorSig     string		// 验证单元签名信息
	Timestamp		time.Time	// 时间戳
	VerifiedUnitNum uint32		// 验证单元编号
}