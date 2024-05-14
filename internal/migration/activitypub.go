package migration

import (
	"fmt"
	"net/url"
)

func creatUserId(name string, domain *url.URL) string {
	return fmt.Sprintf("%s@%s", name, domain.Host)
}

func getInstanceId(name string, domain *url.URL) string {
	return fmt.Sprintf("%s@%s", name, domain.Host)
}
