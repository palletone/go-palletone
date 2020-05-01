// Copyright 2018-2020 PalletOne
// You should have received a copy of the GNU Lesser General Public License

package configure

import (
	"fmt"
)

const (
	VersionMajor = 1         // Major version component of the current release
	VersionMinor = 0         // Minor version component of the current release
	VersionPatch = 7         // Patch version component of the current release
	VersionMeta  = "hotfix1" // Version metadata to append to the version string
)

// Version holds the textual version string.
var Version = func() string {
	v := fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	v += "-" + VersionMeta
	return v
}()

func VersionWithCommit(gitCommit string) string {
	vsn := Version
	if len(gitCommit) >= 8 {
		vsn += "-" + gitCommit[:8]
	}
	return vsn
}
