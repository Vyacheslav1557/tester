package pkg

import "github.com/valkey-io/valkey-go"

func NewValkeyClient(dsn string) (valkey.Client, error) {
	opts, err := valkey.ParseURL(dsn)
	if err != nil {
		return nil, err
	}

	return valkey.NewClient(opts)
}
