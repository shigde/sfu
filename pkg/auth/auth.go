package auth

type AuthConfig struct {
	JWT  JwtConfig `mapstructure:"jwt"`
}

func (a *AuthConfig) GetJwt() *JwtConfig {
	return &a.JWT
}
