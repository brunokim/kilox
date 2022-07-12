package lox

import "fmt"

type object interface {
	get(name Token) any
	set(name Token, value any)
}

type objectState struct {
	fields   map[string]any
	behavior objectBehavior
}

func newObjectState(behavior objectBehavior) objectState {
	return objectState{
		fields:   make(map[string]any),
		behavior: behavior,
	}
}

func (s objectState) get(obj object, name Token) any {
	v, ok := s.fields[name.Lexeme]
	if ok {
		return v
	}
	m, ok := s.behavior.methods[name.Lexeme]
	if ok {
		return m.bind(obj)
	}
	panic(runtimeError{name, fmt.Sprintf("undefined property in %s", obj)})
}

func (s objectState) set(name Token, value any) {
	s.fields[name.Lexeme] = value
}

type objectBehavior struct {
	methods map[string]function
}

func newObjectBehavior() objectBehavior {
	return objectBehavior{
		methods: make(map[string]function),
	}
}

// ----

type metaType struct{}

func (metaType) String() string {
	return "<meta meta>"
}

type metaClass struct {
	name string
	objectBehavior
}

func newMetaClass(name string) metaClass {
	return metaClass{
		name:           name,
		objectBehavior: newObjectBehavior(),
	}
}

func (meta metaClass) String() string {
	return fmt.Sprintf("<meta %s>", meta.name)
}

// ----

type fieldInitializer struct {
	name  Token
	value any
}

type class struct {
	meta             metaClass
	static           objectState
	fieldInits       []fieldInitializer
	instanceBehavior objectBehavior
}

func newClass(meta metaClass) class {
	return class{
		meta:             meta,
		static:           newObjectState(meta.objectBehavior),
		instanceBehavior: newObjectBehavior(),
	}
}

func (cl class) String() string {
	return fmt.Sprintf("<class %s>", cl.meta.name)
}

func (cl class) get(name Token) any {
	return cl.static.get(cl, name)
}

func (cl class) set(name Token, value any) {
	cl.static.set(name, value)
}

func (cl class) Arity() int {
	if init, ok := cl.instanceBehavior.methods["init"]; ok {
		return init.Arity()
	}
	return 0
}

func (cl class) Call(i *Interpreter, args []any) any {
	is := newInstance(cl)
	for _, fieldInit := range cl.fieldInits {
		is.set(fieldInit.name, fieldInit.value)
	}
	if init, ok := cl.instanceBehavior.methods["init"]; ok {
		init.bind(is).Call(i, args)
	}
	return is
}

// ----

type instance struct {
	class class
	state objectState
}

func newInstance(class class) instance {
	return instance{
		class: class,
		state: newObjectState(class.instanceBehavior),
	}
}

func (is instance) String() string {
	return fmt.Sprintf("<instance %s>", is.class.meta.name)
}

func (is instance) get(name Token) any {
	return is.state.get(is, name)
}

func (is instance) set(name Token, value any) {
	is.state.set(name, value)
}
