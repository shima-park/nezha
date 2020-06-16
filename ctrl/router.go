package ctrl

func (ctrl *Ctrl) setRouter() {
	r := ctrl.engine
	r.GET("/pipeline/list", ctrl.listPipelines)
	r.POST("/pipeline/add", ctrl.addPipeline)
	r.GET("/pipeline/ctrl", ctrl.ctrlPipeline)
	r.GET("/pipeline/components", ctrl.listPipelineComponents)
	r.GET("/pipeline/processors", ctrl.listPipelineProcessors)
	r.GET("/pipeline/visualize", ctrl.pipelineVisualize)
	r.GET("/pipeline/vars", ctrl.pipelineVars)
	r.GET("/pipeline/config", ctrl.pipelineConfig)
	r.GET("/pipeline/config/generate", ctrl.generateConfig)

	r.GET("/component/list", ctrl.listComponents)
	r.GET("/component/config", ctrl.componentConfig)

	r.GET("/processor/list", ctrl.listProcessors)
	r.GET("/processor/config", ctrl.processorConfig)

	r.GET("/plugin/list", ctrl.listPlugins)
	r.POST("/plugin/open", ctrl.openPlugin)
}
