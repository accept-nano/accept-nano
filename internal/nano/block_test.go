package nano

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockCreate(t *testing.T) {
	type TestCase struct {
		previous       string
		account        string
		representative string
		balance        string
		link           string
		key            string
		work           string
		hash           string
		block          blockType
	}
	cases := []TestCase{
		// Receive
		{
			"0000000000000000000000000000000000000000000000000000000000000000",
			"nano_3dqpukm9df5dzkq38reosmxwxu3yfphm5hm5ewn5j6axd5pw5gwyqc9dhefx",
			"nano_1ninja7rh37ehfp9utkor5ixmxyg8kme8fnzc4zty145ibch8kf5jwpnzr3r",
			"1000000000000000000000000000000",
			"3910D15E63410512371C3BF4AA7B6C9DD1C5CABC2F9EB0EE153C9214613FFA52",
			"957FE6866F7D0926B5237C8A0A0B332579618B66A39229B487CEDDB2873AAA56",
			"0000000005a6a343",
			"6256BD590ED9662F23D96114739B3A3282530D05F51A5FC53439C8A12178908B",
			blockType{
				Type:           "state",
				Account:        "nano_3dqpukm9df5dzkq38reosmxwxu3yfphm5hm5ewn5j6axd5pw5gwyqc9dhefx",
				Previous:       "0000000000000000000000000000000000000000000000000000000000000000",
				Representative: "nano_1ninja7rh37ehfp9utkor5ixmxyg8kme8fnzc4zty145ibch8kf5jwpnzr3r",
				Balance:        "1000000000000000000000000000000",
				Link:           "3910D15E63410512371C3BF4AA7B6C9DD1C5CABC2F9EB0EE153C9214613FFA52",
				LinkAsAccount:  "nano_1gait7h88ia74aujrgznobxps9gjrq7drdwyp5q3ch6k4jimzykk3wjdpx6u",
				Signature:      "BA8BB46D3C1D9798FEC899E51D1B0AC8CD990EF4D16121DB32447C8761EB15ED2776518E5B0607E5343C1FED848F63F57907917EF52CCCD58D26926150A1DC0D",
				Work:           "0000000005a6a343",
			},
		},
		// Send
		{
			"6256BD590ED9662F23D96114739B3A3282530D05F51A5FC53439C8A12178908B",
			"nano_3dqpukm9df5dzkq38reosmxwxu3yfphm5hm5ewn5j6axd5pw5gwyqc9dhefx",
			"nano_1ninja7rh37ehfp9utkor5ixmxyg8kme8fnzc4zty145ibch8kf5jwpnzr3r",
			"0",
			"nano_1cenk13d5i7qi51ox3m8ipdqecuagopihyb5snffd49b8zo6pq68gqc89nfw",
			"957FE6866F7D0926B5237C8A0A0B332579618B66A39229B487CEDDB2873AAA56",
			"0000000013e65b33",
			"EEDEB24F11A5F3CD0D71C1FF7E0CE442968BE25FAB2D9A88E48B74F67A410651",
			blockType{
				Type:           "state",
				Account:        "nano_3dqpukm9df5dzkq38reosmxwxu3yfphm5hm5ewn5j6axd5pw5gwyqc9dhefx",
				Previous:       "6256BD590ED9662F23D96114739B3A3282530D05F51A5FC53439C8A12178908B",
				Representative: "nano_1ninja7rh37ehfp9utkor5ixmxyg8kme8fnzc4zty145ibch8kf5jwpnzr3r",
				Balance:        "0",
				Link:           "29949002B1C0B780C15E86668597762B68756D07F923CD1AD588E937EA4B5C86",
				LinkAsAccount:  "nano_1cenk13d5i7qi51ox3m8ipdqecuagopihyb5snffd49b8zo6pq68gqc89nfw",
				Signature:      "CC3B8C7774C13A3FD9490A9249AF648F9F25D61AEE3745D80C95D74578E58E6E31549BB80AC4C1F66A3F1B73785A0BAC78B487DCCD0C2504CB7BDFD3EE2C8007",
				Work:           "0000000013e65b33",
			},
		},
	}

	for _, tc := range cases {
		block, err := blockCreate(tc.previous, tc.account, tc.representative, tc.balance, tc.link, tc.key, tc.work)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, tc.hash, block.Hash)
		assert.Equal(t, tc.block, block.Block)
	}
}
