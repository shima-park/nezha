package ctrl

var (
	defaultOptions = Options{
		HTTPAddr: ":8080",
	}
)

type Options struct {
	HTTPAddr     string
	MetadataPath string
}

type Option func(*Options)

func HTTPAddr(addr string) Option {
	return func(o *Options) {
		o.HTTPAddr = addr
	}
}

func MetadataPath(path string) Option {
	return func(o *Options) {
		o.MetadataPath = path
	}
}
