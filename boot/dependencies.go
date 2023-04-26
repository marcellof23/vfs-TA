package boot

type Dependencies struct {
	cfg Config
}

func InitDependencies(cfg Config) (*Dependencies, error) {
	return &Dependencies{
		cfg: cfg,
	}, nil
}

func (d Dependencies) Config() Config {
	return d.cfg
}
