package instance

import (
	"fmt"
	"net/url"
	"strings"
)

func BuildUserId(name string, domain *url.URL) string {
	return fmt.Sprintf("%s@%s", name, domain.Host)
}

func BuildInstanceId(name string, domain *url.URL) string {
	return fmt.Sprintf("%s@%s", name, domain.Host)
}

func SplitUserId(userId string) (string, string) {
	splitUserId := strings.Split(userId, "@")
	if len(splitUserId) > 1 {
		return splitUserId[0], splitUserId[1]
	}
	if len(splitUserId) == 1 {
		return splitUserId[0], ""
	}
	return "", ""
}
