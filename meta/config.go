package meta

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
	"path/filepath"
	//"time"
)

const (
	CSRFOff = iota
	CSRFLoose
	CSRFStrict
)

// Resolves the configuration location from various sources.
//TODO(kkz): Actually implement this properly
func ResolveConfigLocation() string {
	envConfig := os.Getenv("SGG_CONFIG")
	if envConfig != "" {
		return envConfig
	}

	return "./config/app.toml"
}

// Parses many TOML files from a few different locations to create the app config.
//
// Precedence
//
// command-line (on some,)
//
// conf.d in glob order,
//
// root (app.toml,)
//
// Using this, you could set two postgres configs, 00-postgres.toml and
// 01-postgres.toml, and the latter will overwrite any values 00 has, and
// 00 will overwrite any values app.toml has. Exploit this for production config.
func NewConfig(path string) (conf Config) {
	good := false

	// get outer config dir
	d := filepath.Dir(path)

	// check if main config exists
	_, err := os.Stat(path)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			log.Printf("config: base config not found at %s", path)
		} else {
			log.Panicf("config: stat err :: %v", err)
		}
	} else {
		_, err = toml.DecodeFile(path, &conf)
		if err != nil {
			log.Panicf("config: decode err :: %v", err)
		} else {
			// log.Printf("config: loaded root: %s", path)
			good = true
		}
	}

	// check conf.d
	cd, err := os.Stat(d)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			log.Printf("config: %s doesn't exist, skipping")
		} else {
			log.Panicf("config: stat err :: %v", err)
		}
	} else {
		if cd.IsDir() {
			cdf, _ := filepath.Glob(filepath.Join(d, "conf.d", "*.toml"))

			for _, c := range cdf {
				_, err = toml.DecodeFile(c, &conf)
				if err != nil {
					log.Printf("config: decode err (file: %s) :: %v", c, err)
				} else {
					// log.Printf("config: loaded conf.d: %s", c)
					good = true
				}
			}
		} else {
			log.Printf("config: %s isn't a directory, skipping")
		}
	}

	if !good {
		log.Fatal("config: i didn't load any sort of config. exiting.")
	}

	return conf
}

// Configuration structure. Keep this in alphabetical order.
type Config struct {
	Influx     influxConfig
	NATS       natsConfig
	Postgres   pgConfig
	Redis      redisConfig
	Security   securityConfig
	Self       selfConfig
	Validation validationConfig
	Webserver  webserverConfig
}

type influxConfig struct {
	Addr string
	User string
	Pass string
}

type natsConfig struct {
	URL string
}

type pgConfig struct {
	URL string
}

type redisConfig struct {
	Addr string
}

type securityConfig struct {
	SigningKeys struct {
		CSRF     string
		Internal string
		Session  string
	} `toml:"signing_keys"`
}

type selfConfig struct {
	Env           string
	Revision      string
	SessionCookie string `toml:"session_cookies"`
}

type validationConfig struct {
	PasswordLength  int      `toml:"password_length"`
	DisallowedSlugs []string `toml:"disallowed_slugs"`
	UsernameLength  int      `toml:"username_length"`
}

type webserverConfig struct {
	Addr string
	TLS  struct {
		Cert    string
		Private string
	}
}
