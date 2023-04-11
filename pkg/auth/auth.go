package auth

type AuthConfig struct {
	JWT *JwtToken `mapstructure:"jwt"`
}
