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

	// Default authentication config at service level
	DefaultAuth        bool   // Whether authentication is required at service level
	DefaultAuthSchemes string // Default authentication schemes at service level
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
	Num     int    // One RPC method may correspond to multiple HTTP requests
	Request string // SayHelloReq
	Reply   string // SayHelloResp
	// http_rule
	Path         string // HTTP route
	Method       string // HTTP Method
	Body         string
	ResponseBody string
	// Swagger documentation
	Summary     string // API summary
	Description string // API description
	Tags        string // API tags
	Deprecated  bool   // Whether the API is deprecated
	// Authentication
	RequireAuth bool   // Whether this method requires authentication
	AuthSchemes string // Authentication schemes (e.g. BearerAuth, AccessKeyAuth, SecretKeyAuth, etc.)
}

// HandlerName for gin handler name
func (m *method) HandlerName() string {
	return fmt.Sprintf("%s_%d", m.Name, m.Num)
}

// HasPathParams checks if the route contains path parameters
func (m *method) HasPathParams() bool {
	paths := strings.Split(m.Path, "/")
	for _, p := range paths {
		if len(p) > 0 && (p[0] == '{' && p[len(p)-1] == '}' || p[0] == ':') {
			return true
		}
	}
	return false
}

// initPathParams converts route parameters {xx} --> :xx
func (m *method) initPathParams() {
	paths := strings.Split(m.Path, "/")
	for i, p := range paths {
		if len(p) > 0 && (p[0] == '{' && p[len(p)-1] == '}' || p[0] == ':') {
			paths[i] = ":" + p[1:len(p)-1]
		}
	}
	m.Path = strings.Join(paths, "/")
}

// GetSwaggerSummary returns Swagger summary, or default if not set
func (m *method) GetSwaggerSummary() string {
	if m.Summary != "" {
		return m.Summary
	}
	return m.Name
}

// GetSwaggerDescription returns Swagger description, or default if not set
func (m *method) GetSwaggerDescription() string {
	if m.Description != "" {
		return m.Description
	}
	return fmt.Sprintf("%s API endpoint", m.Name)
}

// GetSwaggerTags returns Swagger tags
func (m *method) GetSwaggerTags() string {
	if m.Tags != "" {
		return m.Tags
	}
	return "api"
}

// GetSwaggerMethod returns lowercase HTTP method for Swagger annotation
func (m *method) GetSwaggerMethod() string {
	return strings.ToLower(m.Method)
}

// GetSwaggerPath returns path with leading slash for Swagger annotation and route registration
func (m *method) GetSwaggerPath() string {
	if strings.HasPrefix(m.Path, "/") {
		return m.Path
	}
	return "/" + m.Path
}

// IsQueryMethod checks if the method is a query method (GET, DELETE)
func (m *method) IsQueryMethod() bool {
	return m.Method == "GET" || m.Method == "DELETE"
}

// GetSwaggerParamComment generates parameter comments based on HTTP method
func (m *method) GetSwaggerParamComment() string {
	if m.IsQueryMethod() {
		// For GET and DELETE requests, parameters are passed via query string
		// Should generate comments based on actual request struct fields
		// Temporarily use generic comments, should analyze proto message definition
		return "// @Param request query " + m.Request + " true \"Query parameters\""
	}
	// For POST, PUT, etc. requests, parameters are passed via request body
	return "// @Param request body " + m.Request + " true \"Request body\""
}

// GetAuthComment generates swagger @Security annotation
func (m *method) GetAuthComment() string {
	if !m.RequireAuth || m.AuthSchemes == "" {
		return ""
	}
	var lines []string
	for _, scheme := range strings.Split(m.AuthSchemes, ",") {
		scheme = strings.TrimSpace(scheme)
		if scheme != "" {
			lines = append(lines, "// @Security "+scheme)
		}
	}
	return strings.Join(lines, "\n")
}
