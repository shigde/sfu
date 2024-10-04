package session

import (
	"fmt"
)

type SecurityConfig struct {
	JWT            *JwtToken `mapstructure:"jwt"`
	TrustedOrigins []string  `mapstructure:"trustedOrigins"`
}

func ValidateSecurityConfig(config *SecurityConfig) error {
	if len(config.JWT.Key) < 1 {
		return fmt.Errorf("security.jwt.key should not be empty")
	}

	if len(config.TrustedOrigins) < 1 {
		return fmt.Errorf("security.trustedOrigins should not be empty list")
	}
	return nil
}
