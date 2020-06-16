package ctrl

type Options struct {
	PluginPaths   []string
	PipelinePaths []string
	HTTPAddr      string
}

func (options *Options) Apply(opts ...Option) {
	for _, opt := range opts {
		opt(options)
	}
}

type Option func(*Options)

func PipelinePath(paths ...string) Option {
	return func(o *Options) {
		o.PipelinePaths = append(o.PipelinePaths, paths...)
	}
}

func PluginPath(paths ...string) Option {
	return func(o *Options) {
		o.PluginPaths = append(o.PluginPaths, paths...)
	}
}

func HTTPAddr(addr string) Option {
	return func(o *Options) {
		o.HTTPAddr = addr
	}
}
