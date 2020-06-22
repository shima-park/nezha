package proto

type ControlCommand string

const (
	ControlCommandStart   ControlCommand = "start"
	ControlCommandStop    ControlCommand = "stop"
	ControlCommandRestart ControlCommand = "restart"
)

type VisualizeFormat string

const (
	VisualizeFormatSVG        VisualizeFormat = "svg"
	VisualizeFormatRaw        VisualizeFormat = "raw"
	VisualizeFormatDot        VisualizeFormat = "dot"
	VisualizeFormatASCIITable VisualizeFormat = "ascii_table"
)

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type PipelineView struct {
	Name          string          `json:"name"`
	State         string          `json:"state"`
	Schedule      string          `json:"schedule"`
	Bootstrap     bool            `json:"bootstrap"`
	StartTime     string          `json:"start_time"`
	ExitTime      string          `json:"exit_time"`
	RunTimes      string          `json:"run_times"`
	NextRunTime   string          `json:"next_run_time"`
	LastStartTime string          `json:"last_start_time"`
	LastEndTime   string          `json:"last_end_time"`
	Components    []ComponentView `json:"components,emitempty"`
	Processors    []ProcessorView `json:"processors,emitempty"`
	RawConfig     []byte          `json:"raw_config,emitempty"`
}

type ComponentView struct {
	Name         string `json:"name"`
	RawConfig    string `json:"raw_config,omitempty"`
	SampleConfig string `json:"sample_config"`
	Description  string `json:"description"`
	InjectName   string `json:"inject_name,omitempty"`
	ReflectType  string `json:"reflect_type,omitempty"`
	ReflectValue string `json:"reflect_value,omitempty"`
}

type ProcessorView struct {
	Name         string `json:"name"`
	RawConfig    string `json:"raw_config,omitempty"`
	SampleConfig string `json:"sample_config"`
	Description  string `json:"description"`
}

type PluginView struct {
	Path     string `json:"path"`
	Module   string `json:"module"`
	OpenTime string `json:"open_time"`
}

type MetadataView struct {
	PluginPaths         []string `json:"plugin_paths" yaml:"plugin_paths"`
	PipelineConfigPaths []string `json:"pipeline_config_paths" yaml:"pipeline_config_paths"`
	HTTPAddr            string   `json:"http_addr" yaml:"http_addr"`
}

type PluginOpenRequest struct {
	Path string `json:"path"`
}
