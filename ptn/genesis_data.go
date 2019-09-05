package ptn

var (
	len_data      = 10
	MainNetKeys   [][]byte
	MainNetValues [][]byte
	TestNetKeys   [][]byte
	TestNetValues [][]byte
)

func init() {
	for i := 0; i < len_data; i++ {
		temp_key := make([]byte, 0)
		temp_value := make([]byte, 0)
		MainNetKeys = append(MainNetKeys, temp_key)
		MainNetValues = append(MainNetValues, temp_value)
	}
	for i := 0; i < len_data; i++ {
		temp_key := make([]byte, 0)
		temp_value := make([]byte, 0)
		TestNetKeys = append(TestNetKeys, temp_key)
		TestNetValues = append(TestNetValues, temp_value)
	}
	MainNetKeys[0] = []byte{0x1f}
	MainNetValues[0] = []byte{0x00}

	TestNetKeys[0] = []byte{0x1f}
	TestNetValues[0] = []byte{0x00}

}
