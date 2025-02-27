package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	contextPkg         = protogen.GoImportPath("context")
	ginPkg             = protogen.GoImportPath("github.com/gin-gonic/gin")
	metadataPkg        = protogen.GoImportPath("google.golang.org/grpc/metadata")
	deprecationComment = "// Deprecated: Do not use."
)

var methodSets = make(map[string]int)

func getGoPackageName(gen *protogen.Plugin, goImportPath protogen.GoImportPath) protogen.GoPackageName {
	for _, f := range gen.FilesByPath {
		if f.GoImportPath == goImportPath {
			return f.GoPackageName
		}
	}

	panic(fmt.Errorf("no such package: %v", goImportPath))
}

func getFullTypeName(gen *protogen.Plugin, t *protogen.Message, thisFile *protogen.File) string {
	goImportPath := t.GoIdent.GoImportPath
	if thisFile.GoImportPath == goImportPath {
		return t.GoIdent.GoName
	} else {
		goPackageName := getGoPackageName(gen, goImportPath)
		return fmt.Sprintf("%s.%s", goPackageName, t.GoIdent.GoName)
	}
}

// generateFile generates a _gin.pb.go file.
func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}

	needImport := map[string]struct{}{}
	for _, d := range file.Proto.Dependency {
		df := gen.FilesByPath[d]
		if df != nil {
			needImport[string(df.GoImportPath)] = struct{}{}
		}
	}

	filename := file.GeneratedFilenamePrefix + "_gin.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by github.com/mohuishou/protoc-gen-go-gin. DO NOT EDIT.")
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the protoc-gen-go-gin package it is being compiled against.")
	g.P("import (")
	g.P(contextPkg)
	g.P(ginPkg)
	for pkg, _ := range needImport {
		g.P(protogen.GoImportPath(pkg))
	}
	g.P(metadataPkg)
	g.P(")")
	g.P()

	for _, s := range file.Services {
		genService(gen, file, g, s)
	}
	return g
}

func genService(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, s *protogen.Service) {
	if s.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	// HTTP Server.
	sd := &service{
		Name:     s.GoName,
		FullName: string(s.Desc.FullName()),
		FilePath: file.Desc.Path(),
	}

	for _, m := range s.Methods {
		sd.Methods = append(sd.Methods, genMethod(gen, m, file)...)
	}
	g.P(sd.execute())
}

func genMethod(gen *protogen.Plugin, m *protogen.Method, thisFile *protogen.File) []*method {
	var methods []*method

	// 存在 http rule 配置
	rule, ok := proto.GetExtension(m.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
	if rule != nil && ok {
		for _, bind := range rule.AdditionalBindings {
			methods = append(methods, buildHTTPRule(gen, m, bind, thisFile))
		}
		methods = append(methods, buildHTTPRule(gen, m, rule, thisFile))
		return methods
	}

	// 不存在走默认流程
	methods = append(methods, defaultMethod(gen, m, thisFile))
	return methods
}

// defaultMethodPath 根据函数名生成 http 路由
// 例如: GetBlogArticles ==> get: /blog/articles
// 如果方法名首个单词不是 http method 映射，那么默认返回 POST
func defaultMethod(gen *protogen.Plugin, m *protogen.Method, thisFile *protogen.File) *method {
	names := strings.Split(toSnakeCase(m.GoName), "_")
	var (
		paths      []string
		httpMethod string
		path       string
	)

	switch strings.ToUpper(names[0]) {
	case http.MethodGet, "FIND", "QUERY", "LIST", "SEARCH":
		httpMethod = http.MethodGet
	case http.MethodPost, "CREATE":
		httpMethod = http.MethodPost
	case http.MethodPut, "UPDATE":
		httpMethod = http.MethodPut
	case http.MethodPatch:
		httpMethod = http.MethodPatch
	case http.MethodDelete:
		httpMethod = http.MethodDelete
	default:
		httpMethod = "POST"
		paths = names
	}

	if len(paths) > 0 {
		path = strings.Join(paths, "/")
	}

	if len(names) > 1 {
		path = strings.Join(names[1:], "/")
	}

	md := buildMethodDesc(gen, m, httpMethod, path, thisFile)
	md.Body = "*"
	return md
}

func buildHTTPRule(gen *protogen.Plugin, m *protogen.Method, rule *annotations.HttpRule, thisFile *protogen.File) *method {
	var (
		path   string
		method string
	)
	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = "GET"
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = "PUT"
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = "POST"
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = "DELETE"
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = "PATCH"
	case *annotations.HttpRule_Custom:
		path = pattern.Custom.Path
		method = pattern.Custom.Kind
	}
	md := buildMethodDesc(gen, m, method, path, thisFile)
	return md
}

func buildMethodDesc(gen *protogen.Plugin, m *protogen.Method, httpMethod, path string, thisFile *protogen.File) *method {
	defer func() { methodSets[m.GoName]++ }()
	md := &method{
		Name:    m.GoName,
		Num:     methodSets[m.GoName],
		Request: getFullTypeName(gen, m.Input, thisFile),
		Reply:   getFullTypeName(gen, m.Output, thisFile),
		Path:    path,
		Method:  httpMethod,
	}
	md.initPathParams()
	return md
}

var matchFirstCap = regexp.MustCompile("([A-Z])([A-Z][a-z])")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(input string) string {
	output := matchFirstCap.ReplaceAllString(input, "${1}_${2}")
	output = matchAllCap.ReplaceAllString(output, "${1}_${2}")
	output = strings.ReplaceAll(output, "-", "_")
	return strings.ToLower(output)
}
