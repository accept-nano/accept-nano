package nano

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeterministicKey(t *testing.T) {
	type TestCase struct {
		seed    string
		index   string // must be 32-bit uint but node also accepts 64-bit values
		private string
		public  string
		account string
	}
	cases := []TestCase{
		{
			"0000000000000000000000000000000000000000000000000000000000000000",
			"0",
			"9F0E444C69F77A49BD0BE89DB92C38FE713E0963165CCA12FAF5712D7657120F",
			"C008B814A7D269A1FA3C6528B19201A24D797912DB9996FF02A1FF356E45552B",
			"nano_3i1aq1cchnmbn9x5rsbap8b15akfh7wj7pwskuzi7ahz8oq6cobd99d4r3b7",
		},
		{
			"D32E989C82DAF1DA8F6C358D894C11637DBACC8E234BFC7F6F9FB0F55837A468",
			"111222333",
			"CA554DCDA94A93BC40B2F3F957A1FB65A6842EF70305565A4A45B6E4648D39CE",
			"8FDFDCB5F0849DE82306778B3A8C578F3517DD97602F3CFF69AFB3505BD4F6A4",
			"nano_35yzuktz336xx1jiexwd9c87h5so4zgsgr3h9mzpmdxmc3fxbxo6swb6nmhs",
		},
		{
			"D32E989C82DAF1DA8F6C358D894C11637DBACC8E234BFC7F6F9FB0F55837A468",
			"18446744073709551615", // 2^64-1
			"3EC4035BB5BC06B17BA5159AB555D443715ED5BAFF0C04DCAA76C822E35BD3B6",
			"92AF1B55EBE86E9643300D187B8CEE0926AD999234E7F60003B1C036FA8C3536",
			"nano_36oh5fcyqt5gks3m15arhg8gw4b8opes6f99yr119eg18uxarfbpsxietet6",
		},
	}

	for _, tc := range cases {
		key, err := deterministicKey(tc.seed, tc.index)
		if err != nil {
			t.Fatal(err)
		}
		expected := Key{
			Private: tc.private,
			Public:  tc.public,
			Account: tc.account,
		}
		assert.Equal(t, expected, key)
	}
}
