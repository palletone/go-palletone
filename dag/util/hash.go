package util

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/palletone/go-palletone/dag/modules"
)

func GetJointHash(joint *modules.Joint) string {
	joint_byte, _ := json.Marshal(joint)
	return base64.StdEncoding.EncodeToString([]byte(Sha256(joint_byte)))
}
func GetContentUnitHash(unit *modules.Unit) string {
	if unit.ContentHash != "" {
		return base64.StdEncoding.EncodeToString([]byte(unit.ContentHash))
	}
	// contenthash is null.
	return ""
}
func GetBallHash(unit string, arrParentBalls, arrSkiplistBalls []string, bNonserial bool) string {
	objBall := new(modules.Ball)
	if len(arrParentBalls) > 0 {
		objBall.ParentBalls = arrParentBalls[:]
	}
	if len(arrSkiplistBalls) > 0 {
		objBall.SkiplistBalls = arrSkiplistBalls[:]
	}
	if bNonserial {
		objBall.IsNonserial = true
	}
	ball_byte, _ := json.Marshal(objBall)
	return base64.StdEncoding.EncodeToString([]byte(Sha256(ball_byte)))
}
func Sha1(data []byte) string {
	s1 := sha1.New()
	s1.Write(data)
	return fmt.Sprintf("%x", s1.Sum(nil))
}
func Sha256(data []byte) string {
	s256 := sha256.New()
	s256.Write(data)
	return fmt.Sprintf("%x", s256.Sum(nil))
}
func Sha512(data []byte) string {
	s512 := sha512.New()
	s512.Write(data)
	return fmt.Sprintf("%x", s512.Sum(nil))
}
func Md5(data []byte) string {
	m5 := md5.New()
	m5.Write(data)
	return fmt.Sprintf("%x", m5.Sum(nil))
}
