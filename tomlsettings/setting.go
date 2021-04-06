/*
配置文件读取，持久化，序列化，反序列化
*/
package tomlsettings

import (
	"bytes"
	"sync/atomic"

	"github.com/pelletier/go-toml"
	"github.com/spf13/viper"

	"github.com/cairns0214/go-pkg/file"
)

type Config interface {
	Set(map[string]interface{})
	SetDefault(map[string]interface{})
	Marshal() (string, error)
	Unmarshal(interface{}) error
	Read(fileType string, str string) error
	Write() error
	Load() interface{}
	Store(interface{}) error
	Reset()
	Unset(pkey, key string)
	Delete(key []string) error
	List() map[string]interface{}
}

type config struct {
	filename string
	viper    *viper.Viper
	// content 用来存放配置对应的结构体,
	// 不使用具体的结构体定义
	content atomic.Value
}

func NewEmptyConfig() Config {
	c := &config{
		filename: "",
		viper:    viper.New(),
	}
	return c
}

func NewConfig(filename string) (Config, error) {
	c := &config{
		filename: filename,
	}
	if err := file.AutoCreateFile(filename, 0755); err != nil {
		return c, err
	}
	c.viper = viper.New()
	c.viper.SetConfigFile(filename)
	if err := c.viper.ReadInConfig(); err != nil {
		return c, err
	}
	return c, nil
}

func (c *config) List() map[string]interface{} {
	return c.viper.AllSettings()
}

func (c *config) Reset() {
	c.viper = viper.New()
	c.viper.SetConfigFile(c.filename)
}

func (c *config) Unset(pkey, key string) {
	if nil != c.viper.Get(pkey) {
		delete(c.viper.Get(pkey).(map[string]interface{}), key)
	}
}

func (c *config) Set(iMap map[string]interface{}) {
	for k, v := range iMap {
		c.viper.Set(k, v)
	}
}

// SetDefault 设置配置默认值
func (c *config) SetDefault(iMap map[string]interface{}) {
	for k, v := range iMap {
		c.viper.SetDefault(k, v)
	}
}

func (c *config) Delete(keys []string) error {
	configMap := c.viper.AllSettings()
	for _, k := range keys {
		delete(configMap, k)
	}
	tree, err := toml.TreeFromMap(configMap)
	if err != nil {
		return err
	}
	s, err := tree.ToTomlString()
	if err != nil {
		return err
	}
	c.Reset()
	if err := c.Read("toml", s); err != nil {
		return err
	}
	return nil
}

// Marshal 将内存中的配置，序列化成字符串
func (c *config) Marshal() (string, error) {
	return MarshalWithToml(c.viper)
}

// Unmarshal 将内存中的配置，反序列绑定到结构体
func (c *config) Unmarshal(i interface{}) error {
	if err := c.viper.Unmarshal(i); err != nil {
		return err
	}
	return nil
}

func (c *config) Read(fileType string, str string) error {
	c.viper.SetConfigType(fileType)
	return c.viper.ReadConfig(bytes.NewBufferString(str))
}

func (c config) Write() error {
	return c.viper.WriteConfig()
}

func (c *config) Load() interface{} {
	return c.content.Load()
}

func (c *config) Store(i interface{}) error {
	c.content.Store(i)
	return nil
}

func MarshalWithToml(v *viper.Viper) (string, error) {
	settings := v.AllSettings()
	tree, err := toml.TreeFromMap(settings)
	if err != nil {
		return "", err
	}
	str, err := tree.ToTomlString()
	if err != nil {
		return "", err
	}
	return str, nil
}
