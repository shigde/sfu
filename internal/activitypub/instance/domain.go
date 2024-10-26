package instance

import (
	"fmt"
	"net/url"
)

func BuildUserId(name string, domain *url.URL) string {
	return fmt.Sprintf("%s@%s", name, domain.Host)
}

func BuildInstanceId(name string, domain *url.URL) string {
	return fmt.Sprintf("%s@%s", name, domain.Host)
}
