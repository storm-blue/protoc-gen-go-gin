package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed template.go.tpl
var tpl string

type service struct {
	Name     string // Greeter
	FullName string // helloworld.Greeter
	FilePath string // api/helloworld/helloworld.proto

	Methods   []*method
	MethodSet map[string]*method
}

func (s *service) execute() string {
	if s.MethodSet == nil {
		s.MethodSet = map[string]*method{}
		for _, m := range s.Methods {
			m := m
			s.MethodSet[m.Name] = m
		}
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(tpl))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return buf.String()
}

// InterfaceName service interface name
func (s *service) InterfaceName() string {
	return s.Name + "HTTPServer"
}

type method struct {
	Name    string // SayHello
	Num     int    // 一个 rpc 方法可以对应多个 http 请求
	Request string // SayHelloReq
	Reply   string // SayHelloResp
	// http_rule
	Path         string // 路由
	Method       string // HTTP Method
	Body         string
	ResponseBody string
	// Swagger documentation
	Summary     string // API summary
	Description string // API description
	Tags        string // API tags
	Deprecated  bool   // Whether the API is deprecated
}

// HandlerName for gin handler name
func (m *method) HandlerName() string {
	return fmt.Sprintf("%s_%d", m.Name, m.Num)
}

// HasPathParams 是否包含路由参数
func (m *method) HasPathParams() bool {
	paths := strings.Split(m.Path, "/")
	for _, p := range paths {
		if len(p) > 0 && (p[0] == '{' && p[len(p)-1] == '}' || p[0] == ':') {
			return true
		}
	}
	return false
}

// initPathParams 转换参数路由 {xx} --> :xx
func (m *method) initPathParams() {
	paths := strings.Split(m.Path, "/")
	for i, p := range paths {
		if len(p) > 0 && (p[0] == '{' && p[len(p)-1] == '}' || p[0] == ':') {
			paths[i] = ":" + p[1:len(p)-1]
		}
	}
	m.Path = strings.Join(paths, "/")
}

// GetSwaggerSummary 获取Swagger摘要，如果没有设置则使用默认值
func (m *method) GetSwaggerSummary() string {
	if m.Summary != "" {
		return m.Summary
	}
	return m.Name
}

// GetSwaggerDescription 获取Swagger描述，如果没有设置则使用默认值
func (m *method) GetSwaggerDescription() string {
	if m.Description != "" {
		return m.Description
	}
	return fmt.Sprintf("%s API endpoint", m.Name)
}

// GetSwaggerTags 获取Swagger标签
func (m *method) GetSwaggerTags() string {
	if m.Tags != "" {
		return m.Tags
	}
	return "api"
}

// GetSwaggerMethod 获取小写的HTTP方法，用于Swagger注释
func (m *method) GetSwaggerMethod() string {
	return strings.ToLower(m.Method)
}

// GetSwaggerPath 获取带前导斜杠的路径，用于Swagger注释和路由注册
func (m *method) GetSwaggerPath() string {
	if strings.HasPrefix(m.Path, "/") {
		return m.Path
	}
	return "/" + m.Path
}

// IsQueryMethod 判断是否为查询方法（GET、DELETE）
func (m *method) IsQueryMethod() bool {
	return m.Method == "GET" || m.Method == "DELETE"
}

// GetSwaggerParamComment 根据HTTP方法生成不同的参数注释
func (m *method) GetSwaggerParamComment() string {
	if m.IsQueryMethod() {
		// 对于GET和DELETE请求，参数通过查询字符串传递
		// 这里应该根据实际的请求结构体字段生成注释
		// 暂时使用通用的注释，实际应该分析proto消息定义
		return "// @Param request query " + m.Request + " true \"Query parameters\""
	}
	// 对于POST、PUT等请求，参数通过请求体传递
	return "// @Param request body " + m.Request + " true \"Request body\""
}
