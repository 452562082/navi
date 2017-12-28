package agent

import (
	"git.oschina.net/kuaishangtong/navi/errors"
)

//PluginContainer represents a plugin container that defines all methods to manage plugins.
//And it also defines all extension points.
type PluginContainer interface {
	Add(plugin Plugin)
	Remove(plugin Plugin)
	All() []Plugin

	DoRegister(name string, rcvr interface{}, metadata string) error
	DoUnRegister(name string) error
}

// Plugin is the server plugin interface.
type Plugin interface {
}

type (
	// RegisterPlugin is .
	RegisterPlugin interface {
		Register(name string, rcvr interface{}, metadata string) error
		UnRegister(name string) (err error)
	}
)

// pluginContainer implements PluginContainer interface.
type pluginContainer struct {
	plugins []Plugin
}

// Add adds a plugin.
func (p *pluginContainer) Add(plugin Plugin) {
	p.plugins = append(p.plugins, plugin)
}

// Remove removes a plugin by it's name.
func (p *pluginContainer) Remove(plugin Plugin) {
	if p.plugins == nil {
		return
	}

	var plugins []Plugin
	for _, p := range p.plugins {
		if p != plugin {
			plugins = append(plugins, p)
		}
	}

	p.plugins = plugins
}

func (p *pluginContainer) All() []Plugin {
	return p.plugins
}

// DoRegister invokes DoRegister plugin.
func (p *pluginContainer) DoRegister(name string, rcvr interface{}, metadata string) error {
	var es []error
	for _, rp := range p.plugins {
		if plugin, ok := rp.(RegisterPlugin); ok {
			err := plugin.Register(name, rcvr, metadata)
			if err != nil {
				es = append(es, err)
			}
		}
	}

	if len(es) > 0 {
		return errors.NewMultiError(es)
	}
	return nil
}

func (p *pluginContainer) DoUnRegister(name string) error {
	var es []error
	for _, rp := range p.plugins {
		if plugin, ok := rp.(RegisterPlugin); ok {
			err := plugin.UnRegister(name)
			if err != nil {
				es = append(es, err)
			}
		}
	}

	if len(es) > 0 {
		return errors.NewMultiError(es)
	}
	return nil
}
