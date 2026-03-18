package useragent

import "fmt"

var globalUserAgent = "kmscli"

func RegisterUserAgent(useragent string) {
	globalUserAgent = fmt.Sprintf(globalUserAgent + "/" + useragent)
}
func GetUserAgent() string {
	return globalUserAgent
}
