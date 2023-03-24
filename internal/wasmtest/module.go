package wasmtest

import (
	"github.com/stealthrocket/wazergo"
	"github.com/tetratelabs/wazero/api"
)

// Module is an implementation of wazero's api.Module interface intended to be
// used as a stub in tests. The main feature is the ability to define a memory
// that the module will be exposing as its exported "memory".
type Module struct {
	api.Module // TODO: implement more features of the interface
	name       string
	memory     moduleMemory
}

// ModuleOption represents configuration options for the Module type.
type ModuleOption = wazergo.Option[*Module]

// Memory sets the memory of a Module instance.
func Memory(memory api.Memory) ModuleOption {
	return wazergo.OptionFunc(func(module *Module) { module.memory.Memory = memory })
}

// NewModule constructs a Module instance with the given name and configuration
// options.
func NewModule(name string, opts ...ModuleOption) *Module {
	module := &Module{name: name}
	module.memory.module = module
	wazergo.Configure(module, opts...)
	return module
}

func (mod *Module) Name() string { return mod.name }

func (mod *Module) Memory() api.Memory { return &mod.memory }

func (mod *Module) ExportedMemory(name string) api.Memory {
	switch name {
	case "memory":
		return &mod.memory
	default:
		return nil
	}
}

type moduleMemory struct {
	module *Module
	api.Memory
}

func (mem *moduleMemory) Definition() api.MemoryDefinition {
	return &moduleMemoryDefinition{mem.module, mem.Memory.Definition()}
}

type moduleMemoryDefinition struct {
	module *Module
	api.MemoryDefinition
}

func (def *moduleMemoryDefinition) ModuleName() string {
	return def.module.name
}

func (def *moduleMemoryDefinition) ExportNames() []string {
	return []string{"memory"}
}
