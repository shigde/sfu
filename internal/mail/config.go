package mail

type MailConfig struct {
	Enable   bool   `mapstructure:"enable"`
	From     string `mapstructure:"from"`
	Pass     string `mapstructure:"password"`
	SmtpHost string `mapstructure:"smtpHost"`
	SmtpPort string `mapstructure:"smtpPort"`
}

func ValidateEmailConfig(config *MailConfig) error {
	return nil
}
