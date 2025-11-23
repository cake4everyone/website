package config

import (
	"log"

	"github.com/spf13/viper"
)

// Load loads the given configuration file as the global config. It
// also loads:
//   - the languages from lang.Load() (see cake4everybot/data/lang)
func Load(config string) {
	log.Println("Loading configuration file(s)...")
	log.Printf("Loading config '%s'\n", config)

	viper.AddConfigPath(".")
	viper.SetConfigFile(config)

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Could not load config file '%s': %v", config, err)
	}

	names := viper.GetStringSlice("additionalConfigs")
	for _, n := range names {
		log.Printf("Loading additional config '%s'...\n", n)
		viper.SetConfigFile(n)
		err = viper.MergeInConfig()
		if err != nil {
			log.Printf("Counld not load additional config '%s': %v\n", n, err)
		}
	}

	log.Println("Loaded configuration file(s)!")
}
