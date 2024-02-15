package types

type DBConfig struct {
	Provider string `json:"provider"`
	DBName   string `json:"dbname"`
	Table    string `json:"table"`
	User     string `json:"user"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}
