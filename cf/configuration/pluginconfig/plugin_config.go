package pluginconfig

import (
	"sync"

	"code.cloudfoundry.org/cli/cf/configuration"
)

//go:generate counterfeiter . PluginConfiguration

type PluginConfiguration interface {
	Plugins() map[string]PluginMetadata
	SetPlugin(string, PluginMetadata)
	GetPluginPath() string
	RemovePlugin(string)
	ListCommands() []string
}

type PluginConfig struct {
	mutex      *sync.RWMutex
	initOnce   *sync.Once
	persistor  configuration.Persistor
	onError    func(error)
	data       *PluginData
	pluginPath string
}

func NewPluginConfig(errorHandler func(error), persistor configuration.Persistor, pluginPath string) *PluginConfig {
	return &PluginConfig{
		data:       NewData(),
		mutex:      new(sync.RWMutex),
		initOnce:   new(sync.Once),
		persistor:  persistor,
		onError:    errorHandler,
		pluginPath: pluginPath,
	}
}

func (c *PluginConfig) GetPluginPath() string {
	return c.pluginPath
}

func (c *PluginConfig) Plugins() map[string]PluginMetadata {
	c.read()
	return c.data.Plugins
}

func (c *PluginConfig) SetPlugin(name string, metadata PluginMetadata) {
	if c.data.Plugins == nil {
		c.data.Plugins = make(map[string]PluginMetadata)
	}
	c.write(func() {
		c.data.Plugins[name] = metadata
	})
}

func (c *PluginConfig) RemovePlugin(name string) {
	c.write(func() {
		delete(c.data.Plugins, name)
	})
}

func (c *PluginConfig) ListCommands() []string {
	plugins := c.Plugins()
	allCommands := []string{}

	for _, plugin := range plugins {
		for _, command := range plugin.Commands {
			allCommands = append(allCommands, command.Name)
		}
	}

	return allCommands
}

func (c *PluginConfig) init() {
	//only read from disk if it was never read
	c.initOnce.Do(func() {
		err := c.persistor.Load(c.data)
		if err != nil {
			c.onError(err)
		}
	})
}

func (c *PluginConfig) read() {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	c.init()
}

func (c *PluginConfig) write(cb func()) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.init()

	cb()

	err := c.persistor.Save(c.data)
	if err != nil {
		c.onError(err)
	}
}

func (c *PluginConfig) Close() {
	c.read()
	// perform a read to ensure write lock has been cleared
}
