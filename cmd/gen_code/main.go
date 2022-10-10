// Package main 用于生成服务的基础代码
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	// pb "github.com/ahfuzhang/daily_coding/2022_10_08/pb_extension/github.com/ahfuzhang/daily_coding"
	// proto "github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"google.golang.org/protobuf/types/descriptorpb"
)

var E_HttpPath = &proto.ExtensionDesc{
	ExtendedType:  (*descriptorpb.MethodOptions)(nil),
	ExtensionType: (*string)(nil),
	Field:         51234,
	Name:          "cheaprpc.ahfuzhang.public.http_path",
	Tag:           "bytes,51234,opt,name=http_path",
	Filename:      "extensions.proto",
}

var (
	templatePath = flag.String("template_path", "", "-template_path=/path/to/tpl/files")
	targetPath   = flag.String("target_path", "./", "-target_path=/path/to/save/go/files")
	protoPath    = flag.String("proto_path", "", "-proto_path=/path/to/include/proto/files")
	sourceProto  = flag.String("source_proto", "", "-source_proto=/path/to/one/source/proto/file")
)

func checkArgs() {
	if s, err := os.Stat(*templatePath); os.IsNotExist(err) || !s.IsDir() {
		log.Fatalln("template_path not exists(or not a dir): " + *templatePath)
	}
	if s, err := os.Stat(*targetPath); os.IsNotExist(err) || !s.IsDir() {
		log.Fatalln("target_path not exists(or not a dir): " + *targetPath)
	}
	if s, err := os.Stat(*protoPath); os.IsNotExist(err) || !s.IsDir() {
		log.Fatalln("proto_path not exists(or not a dir): " + *protoPath)
	}
	if s, err := os.Stat(*sourceProto); os.IsNotExist(err) || s.IsDir() {
		log.Fatalln("source_proto not exists(or not a file): " + *sourceProto)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
	checkArgs()
	pbFile := parseFile(protoPath, sourceProto)
	// log.Println(*pbFile.GetFileOptions().GoPackage)
	fullPath := filepath.Join(*targetPath, *pbFile.GetFileOptions().GoPackage)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		log.Println("mkdir:", fullPath)
		if err = os.MkdirAll(fullPath, os.ModePerm); err != nil {
			log.Fatalln("mkdir fail: err=", err.Error())
		}
	}
	goFilePath := filepath.Dir(fullPath)
	makeMainFile(goFilePath, pbFile)
	makeServiceFiles(goFilePath, pbFile)
	makeGoModFile(goFilePath, pbFile)
}

func makeMainFile(goFilePath string, pbFile *desc.FileDescriptor) {
	t := getTemplateByFileName("main.go.tpl")
	mainFile := createFile(filepath.Join(goFilePath, "main.go"))
	err := t.Execute(mainFile, map[string]interface{}{
		"Package":  filepath.Dir(*pbFile.GetFileOptions().GoPackage), // pb 的上一级目录
		"Services": getServiceDirNames(pbFile),
	})
	if err != nil {
		log.Fatalf("write main.go error: err=%s", err.Error())
	}
	mainFile.Close()
}

func makeGoModFile(goFilePath string, pbFile *desc.FileDescriptor) {
	t := getTemplateByFileName("go.mod.tpl")
	goModFile := createFile(filepath.Join(goFilePath, "go.mod"))
	err := t.Execute(goModFile, map[string]string{
		"Package": filepath.Dir(*pbFile.GetFileOptions().GoPackage), // pb 的上一级目录
	})
	if err != nil {
		log.Fatalf("write go.mod error: err=%s", err.Error())
	}
	goModFile.Close()
}

func mkdirs(d string) {
	if _, err := os.Stat(d); os.IsNotExist(err) {
		if err = os.MkdirAll(d, os.ModePerm); err != nil {
			log.Fatalf("create dir [%s] fail, err=%s", d, err.Error())
		}
	}
}

type method struct {
	Path   string
	Method string
	Req    string
	Rsp    string
}

type templateParam struct {
	Import         string
	Service        string
	ServicePackage string
	Methods        []method
}

func createFile(p string) *os.File {
	out, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Fatalf("create file [%s] fail, err=%s", p, err.Error())
	}
	return out
}

func eachMethod(m *desc.MethodDescriptor) *method {
	out := &method{
		Path:   m.GetFullyQualifiedName(),
		Method: m.GetName(),
		Req:    m.GetInputType().GetName(),
		Rsp:    m.GetOutputType().GetName(),
	}
	//
	// if !proto.HasExtension(m.GetMethodOptions(), E_CheapRpcAlias) {
	// 	fmt.Printf("not found ext of method")
	// 	return out
	// }
	// op1, err := proto.GetExtension(m.GetMethodOptions(), E_CheapRpcAlias)
	// if err != nil {
	// 	fmt.Printf("err=%+v", err)
	// 	return out
	// }
	// // fmt.Printf("%+v\n", op1)
	// path1 := op1.(*string)
	// out.Path = *path1
	return out
}

func eachMethodForRegister(m *desc.MethodDescriptor) []method {
	out := []method{
		{
			Path:   "/" + m.GetFullyQualifiedName(),
			Method: m.GetName(),
			Req:    m.GetInputType().GetName(),
			Rsp:    m.GetOutputType().GetName(),
		},
	}

	if !proto.HasExtension(m.GetMethodOptions(), E_HttpPath) {
		return out
	}
	op1, err := proto.GetExtension(m.GetMethodOptions(), E_HttpPath)
	if err != nil {
		return out
	}
	// fmt.Printf("%+v\n", op1)
	httpPath := op1.(*string)
	out = append(out,
		method{
			Path:   *httpPath,
			Method: m.GetName(),
			Req:    m.GetInputType().GetName(),
			Rsp:    m.GetOutputType().GetName(),
		},
	)
	return out
}

func eachService(s *desc.ServiceDescriptor, pbFile *desc.FileDescriptor) *templateParam {
	param := &templateParam{
		Import:         *pbFile.GetFileOptions().GoPackage,
		Service:        s.GetName(),
		ServicePackage: strings.ToLower(s.GetName()),
		Methods:        nil,
	}
	methods := s.GetMethods()
	for _, m := range methods {
		param.Methods = append(param.Methods, *eachMethod(m))
	}
	return param
}

func eachServiceForRegister(s *desc.ServiceDescriptor, pbFile *desc.FileDescriptor) *templateParam {
	param := &templateParam{
		Import:         *pbFile.GetFileOptions().GoPackage,
		Service:        s.GetName(),
		ServicePackage: strings.ToLower(s.GetName()),
		Methods:        nil,
	}
	methods := s.GetMethods()
	for _, m := range methods {
		param.Methods = append(param.Methods, eachMethodForRegister(m)...)
	}
	return param
}

func getServiceDirNames(pbFile *desc.FileDescriptor) []string {
	out := make([]string, 0, len(pbFile.GetServices()))
	for _, service := range pbFile.GetServices() {
		out = append(out, strings.ToLower(service.GetName()))
	}
	return out
}

func makeServiceFile(targetPathOfService string, service *desc.ServiceDescriptor, pbFile *desc.FileDescriptor, t *template.Template) {
	serviceFile := filepath.Join(targetPathOfService, strings.ToLower(service.GetName())+".go")
	f := createFile(serviceFile)
	param := eachService(service, pbFile)
	err := t.Execute(f, param)
	if err != nil {
		log.Fatalln(err)
	}
	f.Close()
}

func makeServiceRegisterFile(targetPathOfService string, service *desc.ServiceDescriptor, pbFile *desc.FileDescriptor, t *template.Template) {
	registerFile := filepath.Join(targetPathOfService, "register.go")
	f := createFile(registerFile)
	param := eachServiceForRegister(service, pbFile)
	err := t.Execute(f, param)
	if err != nil {
		log.Fatalln(err)
	}
	f.Close()
}

func getTemplateByName(name string) *template.Template {
	serviceTemplate, err := ioutil.ReadFile(filepath.Join(*templatePath, name+".go.tpl"))
	if err != nil {
		log.Fatalf("read %s.go.tpl fail, err=%s", name, err.Error())
	}
	t := template.New(name)
	t = template.Must(t.Parse(string(serviceTemplate)))
	return t
}

func getTemplateByFileName(name string) *template.Template {
	templateData, err := ioutil.ReadFile(filepath.Join(*templatePath, name))
	if err != nil {
		log.Fatalf("read %s fail, err=%s", name, err.Error())
	}
	t := template.New(name)
	t = template.Must(t.Parse(string(templateData)))
	return t
}

func makeServiceFiles(goFilePath string, pbFile *desc.FileDescriptor) {
	templateOfService := getTemplateByName("service")
	templateOfRegister := getTemplateByName("register")
	for _, service := range pbFile.GetServices() {
		targetPathOfService := filepath.Join(goFilePath, "internal/services/"+strings.ToLower(service.GetName()))
		mkdirs(targetPathOfService)
		//
		makeServiceFile(targetPathOfService, service, pbFile, templateOfService)
		makeServiceRegisterFile(targetPathOfService, service, pbFile, templateOfRegister)
	}
}

func parseFile(protoPath *string, sourceProto *string) *desc.FileDescriptor {
	p := protoparse.Parser{
		IncludeSourceCodeInfo: true,
		ImportPaths:           []string{*protoPath, "."},
	}
	fds, err := p.ParseFiles(*sourceProto)
	if err != nil {
		log.Fatalf("load file [%s] fail, err=%+v", *sourceProto, err)
	}
	if len(fds) != 1 {
		log.Fatalf("load file [%s] fail, count=%d", *sourceProto, len(fds))
	}
	return fds[0]
}

/*
./main -proto_path=/Users/ahfuzhang/code/ -source_proto=../../examples/my_easy_service.proto -target_path=../../examples/ -template_path=../../templates/
*/
