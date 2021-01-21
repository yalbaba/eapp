package config

type RegistryConfig struct {
	TTl int
}

type Options struct {
	TTl int
}

type Option func(*Options)

func WithTTl(ttl int) Option {
	return func(o *Options) {
		o.TTl = ttl
	}
}
