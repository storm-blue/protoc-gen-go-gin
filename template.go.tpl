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
{{if .Deprecated}}
// @Deprecated
{{end}}
// @Summary {{.GetSwaggerSummary}}
// @Description {{.GetSwaggerDescription}}
// @Tags {{.GetSwaggerTags}}
// @Accept json
// @Produce json
{{.GetSwaggerParamComment}}{{if ne .GetAuthComment ""}}
{{.GetAuthComment}}{{end}}
// @Success 200 {object} {{.Reply}}
// @Failure 400 {object} map[string]interface{} "Invalid parameter"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router {{.GetSwaggerPath}} [{{.GetSwaggerMethod}}]
func (s *{{$.Name}}) {{ .HandlerName }} (ctx *gin.Context) {
	var in {{.Request}}
{{if .HasPathParams }}
	if err := ctx.ShouldBindUri(&in); err != nil {
		s.handler.InvalidParam(ctx, err)
		return
	}
{{end}}
{{if eq .Method "GET" "DELETE" }}
	// 兼容 query 参数驼峰和下划线风格
	{
		camelToSnake := func(s string) string {
			var result []rune
			for i, r := range s {
				if unicode.IsUpper(r) {
					if i > 0 {
						result = append(result, '_')
					}
					result = append(result, unicode.ToLower(r))
				} else {
					result = append(result, r)
				}
			}
			return string(result)
		}
		query := ctx.Request.URL.Query()
		for k, v := range query {
			snake := camelToSnake(k)
			if snake != k {
				if _, exists := query[snake]; !exists {
					query[snake] = v
				}
			}
		}
		ctx.Request.URL.RawQuery = query.Encode()
	}
	if err := ctx.ShouldBindQuery(&in); err != nil {
		s.handler.InvalidParam(ctx, err)
		return
	}
{{else if eq .Method "POST" "PUT" }}
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
	s.router.Handle("{{.Method}}", "{{.GetSwaggerPath}}", s.{{ .HandlerName }})
{{end -}}
}