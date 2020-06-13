package pipeline

import (
	"bytes"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/olekukonko/tablewriter"
)

var (
	visualizers = map[string]Visualizer{
		"svg":         DotVisualizer("svg"),
		"png":         DotVisualizer("png"),
		"dot":         DotGrgphVisualizer,
		"ascii_table": AsciiTableVisualizer,
	}
	supportedVisualizerTypes []string
)

func init() {
	for t := range visualizers {
		supportedVisualizerTypes = append(supportedVisualizerTypes, t)
	}
}

type Visualizer func(w io.Writer, pipeline Pipeliner) error

func DotVisualizer(format string) Visualizer {
	return func(w io.Writer, pipeline Pipeliner) error {
		dotFile, err := ioutil.TempFile("", "dot")
		if err != nil {
			return err
		}
		defer os.Remove(dotFile.Name())
		defer dotFile.Close()

		err = DotGrgphVisualizer(dotFile, pipeline)
		if err != nil {
			return err
		}

		outputFile, err := ioutil.TempFile("", format)
		if err != nil {
			return err
		}
		defer os.Remove(outputFile.Name())
		defer outputFile.Close()

		err = exec.Command("dot", "-T"+format, dotFile.Name(), "-o", outputFile.Name()).Run()
		if err != nil {
			return err
		}

		b, err := ioutil.ReadAll(outputFile)
		if err != nil {
			return err
		}

		_, err = w.Write(b)
		return err
	}
}

func AsciiTableVisualizer(w io.Writer, pipeline Pipeliner) error {
	printPipelineComponents(w, pipeline)
	printPipelineProcessors(w, pipeline)
	return nil
}

func DotGrgphVisualizer(w io.Writer, p Pipeliner) error {
	var buffer bytes.Buffer
	buffer.WriteString("digraph {\n")

	buffer.WriteString(`node [shape=plaintext fontname="Sans serif" fontsize="24"];` + "\n")

	buffer.WriteString(fmt.Sprintf(`%s [ label=<
   <table border="1" cellborder="0" cellspacing="1">`+"\n",
		p.Name(),
	))

	first := true
	p.Monitor().Do(func(namespace string, kv expvar.KeyValue) {
		if namespace != p.Name() {
			return
		}
		if first {
			first = false
			buffer.WriteString("<tr><td align=\"left\"><b>" + p.Name() + "</b></td></tr>\n")
		}

		buffer.WriteString("<tr><td align=\"left\">" + kv.Key + ":" + kv.Value.String() + "</td></tr>\n")
	})
	buffer.WriteString("</table>>];\n")
	buffer.WriteString("\n")

	for _, proc := range p.ListProcessors() {
		buffer.WriteString(fmt.Sprintf(`%s [ label=<
   <table border="1" cellborder="0" cellspacing="1">`+"\n",
			proc.Name,
		))

		first := true
		p.Monitor().Do(func(namespace string, kv expvar.KeyValue) {
			if namespace != proc.Name {
				return
			}
			if first {
				first = false
				buffer.WriteString("<tr><td align=\"left\"><b>" + proc.Name + "</b></td></tr>\n")
			}

			buffer.WriteString("<tr><td align=\"left\">" + kv.Key + ":" + kv.Value.String() + "</td></tr>\n")
		})

		buffer.WriteString("</table>>];\n")
		buffer.WriteString("\n")
	}

	buildRefRalationship(p.GetConfig().Stream, &buffer)

	buffer.WriteString("}")
	_, err := w.Write(buffer.Bytes())
	return err
}

func buildRefRalationship(c StreamConfig, w io.Writer) {
	if c.Name == "" {
		return
	}

	for _, x := range c.Childs {
		_, _ = w.Write([]byte(fmt.Sprintf("  %s %s %s;\n", c.Name, "->", x.Name)))
		buildRefRalationship(x, w)
	}
}

func printPipelineComponents(w io.Writer, p Pipeliner) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{
		"Component Name", "Inject Name", "Reflect Type", "Reflect Value", "Description",
	})
	table.SetRowLine(true)
	for _, c := range p.ListComponents() {
		arr := []string{
			c.Name, c.Component.Instance().Name(), c.Component.Instance().Type().String(),
			c.Component.Instance().Value().String(), c.Factory.Description(),
		}

		table.Rich(arr, []tablewriter.Colors{
			tablewriter.Colors{},
			tablewriter.Colors{tablewriter.Normal, tablewriter.FgGreenColor},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
		})

	}
	table.Render()
}

func printPipelineProcessors(w io.Writer, p Pipeliner) {
	mdeErrs := filterMissingDependencyError(p.CheckDependence())

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Processor Name", //"Config",
		"Request struct name", "Request Field", "Request field type", "Request inject name",
		"Response struct name", "Response Field", "Response field type", "Response inject name"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	for _, p := range p.ListProcessors() {
		requests, responses := getFuncReqAndRespReceptorList(p.Processor)
		var i int
		max := len(requests)
		if max < len(responses) {
			max = len(responses)
		}
		for i < max {
			var arr = []string{p.Name} //p.Processor.RawConfig

			var mdeErr *MissingDependencyError
			var req Receptor
			if i < len(requests) {
				req = requests[i]
				mdeErr = matchError(mdeErrs, req)
			}
			arr = append(arr, req.StructName, req.StructFieldName, req.ReflectType, req.InjectName)

			var resp Receptor
			if i < len(responses) {
				resp = responses[i]
			}
			arr = append(arr, resp.StructName, resp.StructFieldName, resp.ReflectType, resp.InjectName)

			if mdeErr != nil {
				table.Rich(arr, []tablewriter.Colors{
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
					tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
					tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{},
				})
			} else {
				table.Rich(arr, []tablewriter.Colors{
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{tablewriter.Normal, tablewriter.FgCyanColor},
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{},
					tablewriter.Colors{tablewriter.Normal, tablewriter.FgGreenColor},
				})
			}
			i++
		}
	}
	table.Render()
}

func matchError(mdeErrs []MissingDependencyError, r Receptor) *MissingDependencyError {
	for _, mdeErr := range mdeErrs {
		if mdeErr.Field == r.StructFieldName &&
			mdeErr.ReflectType == r.ReflectType &&
			mdeErr.InjectName == r.InjectName {
			return &mdeErr
		}
	}
	return nil
}
