package neural

type AiConfig struct {
	Provider string
	BaseUrl  string `koanf:"base_url"`
	ApiKey   string `koanf:"api_key"`
	Model    string `koanf:"model"`
	Temp     float64
	TopK     int     `koanf:"tok_k"`
	RepPen   float64 `koanf:"rep_pen"`
	MaxTok   int     `koanf:"max_tok"`
	Stop     []string
}
