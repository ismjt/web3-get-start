package system

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type (
	Database struct {
		Dialect string `toml:"dialect"`
		DSN     string `toml:"dsn"`
	}

	JWT struct {
		SK     string `toml:"sk"`
		Issuer string `toml:"issuer"`
	}

	Author struct {
		Name  string `toml:"name"`
		Email string `toml:"email"`
	}

	Navigator struct {
		Title  string `toml:"title"`
		Url    string `toml:"url"`
		Target string `toml:"target"`
	}

	Configuration struct {
		Addr          string      `toml:"addr"`
		Title         string      `toml:"title"`
		SessionSecret string      `toml:"session_secret"`
		Domain        string      `toml:"domain"`
		FileServer    string      `toml:"file_server"`
		NotifyEmails  string      `toml:"notify_emails"`
		PageSize      int         `toml:"page_size"`
		PublicDir     string      `toml:"public"`
		ViewDir       string      `toml:"views"`
		Database      Database    `toml:"database"`
		Navigators    []Navigator `toml:"navigators"`
		JWT           JWT         `toml:"jwt"`
	}
)

func (a Author) String() string {
	return fmt.Sprintf("%s,%s", a.Name, a.Email)
}

var configuration *Configuration

func defaultConfig() Configuration {
	return Configuration{
		JWT: JWT{
			Issuer: "personal-blog-server",
			SK:     "776df678g6hd78f6g8h7df8gdh",
		},
		Addr:          ":8090",
		SessionSecret: "asdf89sd7f98a9sd8f78asd",
		Domain:        "https://ismjt.com",
		Title:         "Personal blog",
		FileServer:    "local",
		PageSize:      10,
		PublicDir:     "static",
		ViewDir:       "views/**/*",
		Database: Database{
			Dialect: "sqlite",
			DSN:     "personal_blog.db",
		},
		Navigators: []Navigator{
			{
				Title: "Posts",
				Url:   "/index",
			},
		},
	}
}

func LoadConfiguration(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var config = defaultConfig()
	err = toml.Unmarshal(data, &config)
	if err != nil {
		return err
	}
	configuration = &config
	return nil
}

func Generate() error {
	config := defaultConfig()
	placeholder := "[!!]"
	config.Domain = placeholder
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile("conf/conf.sample.toml", data, os.ModePerm)
}

func GetConfiguration() *Configuration {
	return configuration
}
