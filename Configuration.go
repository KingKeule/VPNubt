package main

// Varaible fields name must start with capital letters to be JSON exported. "dstIP" is not possbile.
type Config struct {
	DstIP   string `json:"IP"`
	SrcPort int    `json:"Port"`
}

func getDefaultConf() *Config {
	config := Config{"0.0.0.0", 0}
	return &config
}

func getWar3Conf() *Config {
	config := Config{"", 6112}
	return &config
}

func getCoDUOConf() *Config {
	config := Config{"", 28960}
	return &config
}
