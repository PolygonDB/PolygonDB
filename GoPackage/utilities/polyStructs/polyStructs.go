package Polystructs

type Config struct {
	Key string `json:"key"`
	Enc bool   `json:"encrypted"`
}

// Settings.json parsing
type Settings struct {
	Addr     string        `json:"addr"`
	Port     string        `json:"port"`
	Logb     bool          `json:"log"`
	Whiteadd []interface{} `json:"whitelist_addresses"`
}
