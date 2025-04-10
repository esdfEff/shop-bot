package config

type Config struct {
	ClientBotToken string
	AdminBotToken  string
	CryptoPayToken string // Добавьте это поле
}

func LoadConfig() Config {
	return Config{
		ClientBotToken: "7041314383:AAFnir8MMWRpWd-9FGmle8N1szYlffxUCfQ",
		AdminBotToken:  "7634372600:AAGuPPeTf-JWQnxsI1GaifcIt-e6-lM1-hI",
		CryptoPayToken: "368429:AAJOnB8gkYvfhqHBa4AOqBmBpKa8UbTTd8E", // Замените на ваш токен Crypto Pay API
	}
}
