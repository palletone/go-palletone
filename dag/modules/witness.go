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

package modules

type WitnessList struct {
	Unit string `json:"unit"`
}

type EarnedHeadersCommissionRecipient struct {
	Unit string `json:"unit"`
}

/* SELECT MAX(units.level) AS max_alt_level \n\
FROM units \n\
LEFT JOIN parenthoods ON units.unit=child_unit \n\
LEFT JOIN units AS punits ON parent_unit=punits.unit AND punits.witnessed_level >= units.witnessed_level \n\
WHERE units.unit IN(?) AND punits.unit IS NULL AND ( \n\
	SELECT COUNT(*) \n\
	FROM unit_witnesses \n\
	WHERE unit_witnesses.unit IN(units.unit, units.witness_list_unit) AND unit_witnesses.address IN(?) \n\
)>=?",  */
type UnitWitness struct {
	Unit    string `json:"unit"`
	Address string `json:"address"`
}

func GetWitnessList() *WitnessList {
	w := new(WitnessList)
	return w

}
func GetUnitWitness() *UnitWitness {
	u := new(UnitWitness)
	return u
}
