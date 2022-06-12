package telemetry

type ServiceOption func(*serviceOptions) error

type serviceOptions struct {
	enableBandwidth      bool
	enableDebug          bool
	defaultStreamOptions []StreamOption
}

func serviceDefaults() *serviceOptions {
	return &serviceOptions{
		enableBandwidth:      true,
		enableDebug:          true,
		defaultStreamOptions: []StreamOption{},
	}
}

func serviceApply(o *serviceOptions, os ...ServiceOption) error {
	for _, opt := range os {
		err := opt(o)
		if err != nil {
			return err
		}
	}
	return nil
}

func WithServiceBandwidth(enabled bool) ServiceOption {
	return func(so *serviceOptions) error {
		so.enableBandwidth = enabled
		return nil
	}
}

func WithServiceDebug(enabled bool) ServiceOption {
	return func(so *serviceOptions) error {
		so.enableDebug = enabled
		return nil
	}
}

func WithServiceDefaultStreamOpts(opts ...StreamOption) ServiceOption {
	return func(so *serviceOptions) error {
		so.defaultStreamOptions = opts
		return nil
	}
}
