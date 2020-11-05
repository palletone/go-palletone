package pdai

import (
	"fmt"
	"github.com/palletone/go-palletone/dag/modules"
	"strings"
	"testing"
)

func TestAmount(t *testing.T) {
	pntAsset := modules.NewPTNAsset()
	amount := pntAsset.DisplayAmount(1000)
	fmt.Println(amount.String())

	assetID, _ := modules.NewAssetId(Dai_symbol, modules.AssetType_FungibleToken,
		byte(Dai_decimal), []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		modules.UniqueIdType_Null)
	wantAmount := assetID.ToAsset().DisplayAmount(1000)
	fmt.Println(wantAmount.String())

	u64Amount := assetID.ToAsset().Uint64Amount(wantAmount)
	fmt.Println(u64Amount)

	str := "1000-1500"
	amountsStr := strings.Split(str, "-")
	fmt.Println(amountsStr)
}
