type {{ $.InterfaceName }} interface {
{{range .MethodSet -}}
	{{.Name}}(context.Context, *{{.Request}}) (*{{.Reply}}, error)
{{end -}}
}

func Register{{ $.InterfaceName }}(r gin.IRouter, srv {{ $.InterfaceName }}) {
	s := {{.Name}}{
		server: srv,
		router:     r,
		handler: _{{$.Name}}Responder{},
	}
	s.RegisterService()
}

type {{$.Name}} struct{
	server {{ $.InterfaceName }}
	router gin.IRouter
	handler  interface {
		Error(ctx *gin.Context, err error)
		InvalidParam(ctx *gin.Context, err error)
		Success(ctx *gin.Context, data interface{})
	}
}

type _{{$.Name}}Responder struct {}

func (handler _{{$.Name}}Responder) sendResponse(ctx *gin.Context, status int, message string, data interface{}) {
	response := map[string]interface{}{}
	if message != "" {
		response["message"] = message
	}
	if data != nil {
		response["data"] = data
	}
	ctx.JSON(status, response)
}

func (handler _{{$.Name}}Responder) Error(ctx *gin.Context, err error) {
	status := 500
	message := "unknown error"
	
	if err != nil {
		message = err.Error()
	}

	_ = ctx.Error(err)
	handler.sendResponse(ctx, status, message, nil)
}

func (handler _{{$.Name}}Responder) InvalidParam(ctx *gin.Context, err error) {
	_ = ctx.Error(err)
	handler.sendResponse(ctx, 400, "invalid param", nil)
}

func (handler _{{$.Name}}Responder) Success(ctx *gin.Context, data interface{}) {
	handler.sendResponse(ctx, 200, "success", data)
}

{{range .Methods}}
func (s *{{$.Name}}) {{ .HandlerName }} (ctx *gin.Context) {
	var in {{.Request}}
{{if .HasPathParams }}
	if err := ctx.ShouldBindUri(&in); err != nil {
		s.handler.InvalidParam(ctx, err)
		return
	}
{{end}}
{{if eq .Method "GET" }}
	if err := ctx.ShouldBindQuery(&in); err != nil {
		s.handler.InvalidParam(ctx, err)
		return
	}
{{else if eq .Method "POST" "PUT" "DELETE" }}
	if err := ctx.ShouldBindJSON(&in); err != nil {
		s.handler.InvalidParam(ctx, err)
		return
	}
{{else}}
	if err := ctx.ShouldBind(&in); err != nil {
		s.handler.InvalidParam(ctx, err)
		return
	}
{{end}}
	md := metadata.New(nil)
	for k, v := range ctx.Request.Header {
		md.Set(k, v...)
	}
	newCtx := metadata.NewIncomingContext(ctx.Request.Context(), md)
	out, err := s.server.({{ $.InterfaceName }}).{{.Name}}(newCtx, &in)
	if err != nil {
		s.handler.Error(ctx, err)
		return
	}

	s.handler.Success(ctx, out)
}
{{end}}

func (s *{{$.Name}}) RegisterService() {
{{range .Methods -}}
	s.router.Handle("{{.Method}}", "{{.Path}}", s.{{ .HandlerName }})
{{end -}}
}