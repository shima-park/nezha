package server

import (
	"io"
	"reflect"

	"github.com/olekukonko/tablewriter"
	"github.com/shima-park/lotus/common/inject"
	"github.com/shima-park/lotus/pipeline"
	"github.com/shima-park/lotus/processor"
)

func init() {
	pipeline.AddVisualizer("ascii_table", ASCIITableVisualizer)
}

func ASCIITableVisualizer(w io.Writer, pipeline pipeline.Pipeliner) error {
	printPipelineComponents(w, pipeline)
	printPipelineProcessors(w, pipeline)
	return nil
}

func printPipelineComponents(w io.Writer, p pipeline.Pipeliner) {
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

func printPipelineProcessors(w io.Writer, p pipeline.Pipeliner) {
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

			var mdeErr *pipeline.MissingDependencyError
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

func filterMissingDependencyError(errs []error) []pipeline.MissingDependencyError {
	var mdeErrs []pipeline.MissingDependencyError
	for _, err := range errs {
		cause, ok := err.(pipeline.MissingDependencyError)
		if ok {
			mdeErrs = append(mdeErrs, cause)
		}
	}
	return mdeErrs
}

func matchError(mdeErrs []pipeline.MissingDependencyError, r Receptor) *pipeline.MissingDependencyError {
	for _, mdeErr := range mdeErrs {
		if mdeErr.Field == r.StructFieldName &&
			mdeErr.ReflectType == r.ReflectType &&
			mdeErr.InjectName == r.InjectName {
			return &mdeErr
		}
	}
	return nil
}

type Receptor struct {
	StructName      string
	StructFieldName string
	InjectName      string
	ReflectType     string
}

func getFuncReqAndRespReceptorList(f interface{}) ([]Receptor, []Receptor) {
	if err := processor.Validate(f); err != nil {
		return nil, nil
	}

	t := reflect.TypeOf(f)

	var reqReceptors []Receptor
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)

		for argType.Kind() == reflect.Ptr {
			argType = argType.Elem()
		}

		if argType.Kind() != reflect.Struct {
			continue
		}

		val := reflect.New(argType)

		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() != reflect.Struct {
			continue
		}

		typ := val.Type()

		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			structField := typ.Field(i)
			injectName := structField.Tag.Get("inject")

			var tt reflect.Type
			if f.Type().Kind() == reflect.Interface {
				nilPtr := reflect.New(f.Type())
				tt = inject.InterfaceOf(nilPtr.Interface())
			} else {
				tt = f.Type()
			}

			reqReceptors = append(reqReceptors, Receptor{
				StructName:      typ.Name(),
				StructFieldName: structField.Name,
				InjectName:      injectName,
				ReflectType:     tt.String(),
			})
		}
	}

	var respReceptors []Receptor
	for i := 0; i < t.NumOut(); i++ {
		outType := t.Out(i)

		if outType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			continue
		}

		for outType.Kind() == reflect.Ptr {
			outType = outType.Elem()
		}

		if outType.Kind() != reflect.Struct {
			continue
		}

		val := reflect.New(outType)
		for val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() != reflect.Struct {
			continue
		}

		typ := val.Type()

		for i := 0; i < val.NumField(); i++ {
			f := val.Field(i)
			structField := typ.Field(i)
			injectName := structField.Tag.Get("inject")

			var tt reflect.Type
			if f.Type().Kind() == reflect.Interface {
				nilPtr := reflect.New(f.Type())
				tt = inject.InterfaceOf(nilPtr.Interface())
			} else {
				tt = f.Type()
			}

			respReceptors = append(respReceptors, Receptor{
				StructName:      typ.Name(),
				StructFieldName: structField.Name,
				InjectName:      injectName,
				ReflectType:     tt.String(),
			})
		}
	}
	return reqReceptors, respReceptors
}
