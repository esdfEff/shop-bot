package config

type Config struct {
	ClientBotToken string
	AdminBotToken  string
}

func LoadConfig() Config {
	return Config{
		ClientBotToken: "7041314383:AAFnir8MMWRpWd-9FGmle8N1szYlffxUCfQ",
		AdminBotToken:  "7634372600:AAGuPPeTf-JWQnxsI1GaifcIt-e6-lM1-hI",
	}
}
