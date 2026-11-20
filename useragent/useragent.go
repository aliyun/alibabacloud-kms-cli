package useragent

const DefaultUserAgent = "kmscli"

var globalUserAgent = DefaultUserAgent

func RegisterUserAgent(useragent string) {
	globalUserAgent = globalUserAgent + "/" + useragent
}
func GetUserAgent() string {
	return globalUserAgent
}
