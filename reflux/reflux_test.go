package reflux

const (
	testPrivFilePath = "../testdata/test.key"
	testPubFilePath  = "../testdata/test.pub"
)

type testJson struct {
	Id        int64  `json:"id"`
	OpenId    string `json:"openId"`
	ExpiredAt int64  `json:"expiredAt"`
	Code      string `json:"code"`
}
