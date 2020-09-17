package nano

import "testing"

// func TestGenerateWork(t *testing.T) {
// 	hash := "4B82B23F17BFC0C6109543962E9C4AFEC8CBBCA6579C438B23B62E6BDCE6E78B"
// 	work, _ := GenerateWork(hash)
// 	if work != "0000000003e8fc75" {
// 		t.FailNow()
// 	}
// }

func BenchmarkGenerateWork(b *testing.B) {
	workThresholdForSend = 0xfff0000000000000
	for n := 0; n < b.N; n++ {
		GenerateWork("42473809202D318F7DFA794B277E78AE3824836A674B1B88BEC1BBC277A87D52", true)
	}
}
