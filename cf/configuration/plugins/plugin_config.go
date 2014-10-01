package configuration

import "sync"

type PluginConfig struct {
	mutex     *sync.RWMutex
	initOnce  *sync.Once
	persistor Persistor
	onError   func(error)
	data      *PluginData
}

func NewPluginConfig(errorHandler func(error)) *PluginConfig {
	c := new(PluginConfig)
	c.mutex = new(sync.RWMutex)
	c.initOnce = new(sync.Once)
	c.persistor = NewDiskPersistor(userHomeDir(), c.data)
	c.onError = errorHandler
	return c
}

func (c *PluginConfig) init() {
	c.initOnce.Do(func() {
		var err error

		data, err := c.persistor.Load()
		if err != nil {
			c.onError(err)
		}

		c.data = data.(*PluginData)

		/*c.data, err = c.persistor.Load()
		if err != nil {
			c.onError(err)
		}*/
	})
}

func (c *PluginConfig) read(cb func()) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	c.init()

	cb()
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

// CLOSERS

func (c *PluginConfig) Close() {
	c.read(func() {
		// perform a read to ensure write lock has been cleared
	})
}
