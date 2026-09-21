package main

import (
	"google.golang.org/protobuf/compiler/protogen"
)

func execute(g *protogen.GeneratedFile, s *serviceDesc) error {
	// pattern constants
	for _, m := range s.Methods {
		g.P("const ", patternConstant(s.ServiceType, m.Name), ` = "`, m.Pattern, `"`)
	}
	g.P()

	// server impl
	{
		// server interface
		g.P("type ", serverInterfaceName(s.ServiceType), " interface {")
		for _, m := range s.Methods {
			g.P(m.Comment)
			g.P(serverMethodDefinition(g, m))
		}
		g.P("}")
		g.P()
		// server factory
		g.P("func Register", s.ServiceType, "AsynqHandler(mux *", g.QualifiedGoIdent(asynqPackage.Ident("ServeMux")), ", srv ", serverInterfaceName(s.ServiceType), ") {")
		for _, m := range s.Methods {
			g.P("mux.HandleFunc(", patternConstant(s.ServiceType, m.Name), ", ", serviceHandlerMethodName(s.ServiceType, m.Name), "(srv))")
		}
		g.P("}")
		g.P()

		// server handler
		for _, m := range s.Methods {
			g.P("func ", serviceHandlerMethodName(s.ServiceType, m.Name), "(srv ", serverInterfaceName(s.ServiceType), ") ", asynqHandler(g, true), " {")
			{ // closure
				g.P("return ", asynqHandler(g, false), " {")
				g.P("var in ", m.Request)
				g.P()
				g.P("if err := ", g.QualifiedGoIdent(protoPackage.Ident("Unmarshal")), "(task.Payload(), &in); err != nil {")
				g.P(`return fmt.Errorf("unmarshal payload failed, %w, %w", err, asynq.SkipRetry)`)
				g.P("}")
				g.P("return srv.", m.Name, "(ctx, &in)")
				g.P("}")
			}
			g.P("}")
			g.P()
		}
	}
	g.P()
	//  client impl
	{
		// client interface
		g.P("type ", clientInterfaceName(s.ServiceType), " interface {")
		for _, m := range s.Methods {
			g.P(m.Comment)
			g.P(clientMethodDefinition(g, m, true))
		}
		g.P("}")
		g.P()

		// client impl
		g.P("type ", clientImplStructName(s.ServiceType), " struct {")
		g.P("cc *", g.QualifiedGoIdent(asynqPackage.Ident("Client")))
		g.P("}")
		g.P()
		// client factory
		g.P("// ", clientFactoryMethodName(s.ServiceType), " new client.")
		g.P("func ", clientFactoryMethodName(s.ServiceType), " (client *", g.QualifiedGoIdent(asynqPackage.Ident("Client")), ") ", clientInterfaceName(s.ServiceType), " {")
		{ // closure
			g.P("return &", clientImplStructName(s.ServiceType), " {")
			g.P("cc: client,")
			g.P("}")
		}
		g.P("}")
		g.P()
		// client method
		for _, m := range s.Methods {
			g.P(m.Comment)
			g.P("func (c *", clientImplStructName(s.ServiceType), ")", clientMethodDefinition(g, m, false), " {")
			if !args.DisableValidate {
				g.P("if err := ", g.QualifiedGoIdent(protovalidatePackage.Ident("Validate")), "(in); err != nil {")
				g.P("return nil, err")
				g.P("}")
			}
			g.P("payload, err := ", g.QualifiedGoIdent(protoPackage.Ident("Marshal")), "(in)")
			g.P("if err != nil {")
			g.P("return nil, err")
			g.P("}")
			g.P("task := ", g.QualifiedGoIdent(asynqPackage.Ident("NewTask")), "(", patternConstant(s.ServiceType, m.Name), ", payload, opts...)")
			g.P("return c.cc.Enqueue(task)")
			g.P("}")
			g.P()
		}
	}
	return nil
}

func patternConstant(serviceType, name string) string {
	return "Pattern_" + serviceType + "_" + name
}

func cronSpecConstant(serviceType, name string) string {
	return "CronSpec_" + serviceType + "_" + name
}

func serverInterfaceName(serviceType string) string {
	return serviceType + "AsynqHandler"
}
func serverMethodDefinition(g *protogen.GeneratedFile, m *methodDesc) string {
	return m.Name + "(" + g.QualifiedGoIdent(contextPackage.Ident("Context")) + ", *" + m.Request + ") error"
}
func serviceHandlerMethodName(serviceType, name string) string {
	return "_" + serviceType + "_" + name + "_Asynq_Handler"
}

func clientInterfaceName(serviceType string) string {
	return serviceType + "AsynqClient"
}
func clientMethodDefinition(g *protogen.GeneratedFile, m *methodDesc, isDeclaration bool) string {
	ctxParam := ""
	inParam := ""
	optsParam := ""
	if !isDeclaration {
		ctxParam = "ctx"
		inParam = "in"
		optsParam = "opts"
	}
	return m.Name + "(" + ctxParam + " " +
		g.QualifiedGoIdent(contextPackage.Ident("Context")) +
		", " + inParam + " *" + m.Request +
		", " + optsParam + " ..." + g.QualifiedGoIdent(asynqPackage.Ident("Option")) +
		") (*" + g.QualifiedGoIdent(asynqPackage.Ident("TaskInfo")) + ", error)"
}
func clientImplStructName(serviceType string) string {
	return serviceType + "AsynqClientImpl"
}
func clientFactoryMethodName(serviceType string) string {
	return "New" + serviceType + "AsynqClient"
}

func asynqHandler(g *protogen.GeneratedFile, isDeclaration bool) string {
	ctxParam := ""
	taskParam := ""
	if !isDeclaration {
		ctxParam = "ctx"
		taskParam = "task"
	}
	return "func(" + ctxParam + " " + g.QualifiedGoIdent(contextPackage.Ident("Context")) + ", " +
		taskParam + " *" + g.QualifiedGoIdent(asynqPackage.Ident("Task")) + ") error"
}
