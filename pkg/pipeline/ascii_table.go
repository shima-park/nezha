package pipeline

import (
	"io"

	"github.com/olekukonko/tablewriter"
)

func PrintPipelineComponents(w io.Writer, p *Pipeline) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{
		"Component Name", "Inject Name", "Reflect Type", "Reflect Value",
		"Raw Config", "Sample Config", "Description",
	})
	table.SetRowLine(true)
	for _, c := range p.ListComponent() {
		arr := []string{
			c.Name, c.InjectName, c.ReflectType, c.ReflectValue,
			c.RawConfig, c.SampleConfig, c.Description,
		}

		table.Rich(arr, []tablewriter.Colors{
			tablewriter.Colors{},
			tablewriter.Colors{tablewriter.Normal, tablewriter.FgGreenColor},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
			tablewriter.Colors{},
		})

	}
	table.Render()
}

func PrintPipelineProcessor(w io.Writer, p *Pipeline) {
	mdeErrs := filterMissingDependencyError(p.CheckDependence())

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Processor Name", "Config",
		"Request struct name", "Request Field", "Request field type", "Request inject name",
		"Response struct name", "Response Field", "Response field type", "Response inject name"})
	table.SetAutoMergeCells(true)
	table.SetRowLine(true)
	for _, p := range p.ListProcessor() {
		var i int
		max := len(p.Request)
		if max < len(p.Response) {
			max = len(p.Response)
		}
		for i < max {
			var arr = []string{p.ProcessorName, p.ProcessorConfig}
			var mdeErr *MissingDependencyError
			var req Receptor
			if i < len(p.Request) {
				req = p.Request[i]
				mdeErr = matchError(mdeErrs, req)
			}
			arr = append(arr, req.StructName, req.StructFieldName, req.ReflectType, req.InjectName)

			var resp Receptor
			if i < len(p.Response) {
				resp = p.Response[i]
			}
			arr = append(arr, resp.StructName, resp.StructFieldName, resp.ReflectType, resp.InjectName)

			if mdeErr != nil {
				table.Rich(arr, []tablewriter.Colors{
					tablewriter.Colors{},
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
