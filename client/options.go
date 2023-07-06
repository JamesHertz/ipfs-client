package client

import recs "github.com/JamesHertz/webmaster/record"

type Config struct {
	apiUrl          string
	mode            recs.IpfsMode
	bootstrapNodes []string
	shouldBootstrap bool
}

type Option func(*Config) error

func defaultConfig() *Config {
	return &Config{
		apiUrl: "localhost:5001",
		mode:   recs.NONE,
	}
}

func (cfg *Config) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return err
		}
	}
	return nil
}

func Api(url string) Option {
	return func(c *Config) error {
		c.apiUrl = url
		return nil
	}
}

func Mode(mode recs.IpfsMode) Option {
	return func(c *Config) error {
		if err := mode.Validate(); err != nil {
			return err
		}
		c.mode = mode
		return nil
	}
}

func Bootstrap(nodes ...string) Option {
	return func(c *Config) error {
		c.bootstrapNodes = nodes
		return nil
	}
}
