package confd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"path"

	"github.com/mkideal/log"
	"github.com/mkideal/pkg/textutil"
	robfigconf "github.com/robfig/config"
)

/*
os.agrv "{\"prefix\":\"\",\"tmpl_src\":\"\",\"toml_dest\":\"\",\"conf_dest\":\"\",\"conf_path\":\"basic.conf\",\"reload_shell\":\"\"}"
*/
const confd_toml_T = `#general by zgen
[template]
prefix = "{{.prefix}}"   #此配置对应etcd的目录
src = "{{.tmpl_src}}"    #对应templates下的模板文件
dest = "{{.conf_dest}}"  #对应服务的配置文件，可以指定目录
keys = [
  {{.keys}}
]
reload_cmd = "{{.reload_shell}}"       #配置文件没问题后的reload
`

type ConfdTomlT struct {
	Prefix      string `json:"prefix"`       //default ""
	TmplSrc     string `json:"tmpl_src"`     //default ./Prefix_config.conf.tmpl
	TomlDest    string `json:"toml_dest"`    //default ./Prefix_config.conf.toml
	ConfDest    string `json:"conf_dest"`    // default ./config.conf
	ConfPath    string `json:"conf_path"`    //require
	ReloadShell string `json:"reload_shell"` //default sh reload.sh
}

const (
	DIR_CONFD     = "conf.d"
	DIR_TEMPLATES = "templates"
)

/*
we have two way for read conf
1、by github.com/robfig/config package. ok
2、by file read then replace
*/
func (c *ConfdTomlT) ConfdTomlGen() error {
	if c.ConfPath == "" {
		return errors.New("confdTomlGen confPath is nil")
	}
	log.Info("current:%v", c)
	conf, err := robfigconf.ReadDefault(c.ConfPath)
	if err != nil {
		return err
	}
	keys := make([]string, 0)
	for _, section := range conf.Sections() {
		options, err := conf.SectionOptions(section)
		if err != nil {
			log.Warn("conf.SectionOptions(%s),err:%v", section, err)
			continue
		}
		for _, option := range options {
			conf.AddOption(section, option, fmt.Sprintf("{{getv \"/%s:%s\"}}", section, option))
			keys = append(keys, fmt.Sprintf("\"/%s:%s\",", section, option))
		}
	}
	os.Mkdir(DIR_CONFD, 0766)

	os.Mkdir(DIR_TEMPLATES, 0766)

	_, file := path.Split(c.ConfPath)
	if c.TmplSrc == "" {
		c.TmplSrc = fmt.Sprintf("%s.tmpl", file)
		if c.Prefix != "" {
			c.TmplSrc = c.Prefix + "_" + c.TmplSrc
		}
	}
	if c.TomlDest == "" {
		c.TomlDest = fmt.Sprintf("%s.toml", file)
		if c.Prefix != "" {
			c.TomlDest = c.Prefix + "_" + c.TomlDest
		}
	}
	c.TomlDest = path.Join(DIR_CONFD, c.TomlDest)
	if c.Prefix != "" {
		c.Prefix = "/" + c.Prefix
	}
	if c.ConfDest == "" {
		c.ConfDest = file
	}
	if c.ReloadShell == "" {
		c.ReloadShell = "sh reload.sh"
	}
	tomlStr := textutil.Tpl(confd_toml_T, map[string]string{
		"prefix":       c.Prefix,
		"tmpl_src":     c.TmplSrc,
		"conf_dest":    c.ConfDest,
		"keys":         strings.Join(keys, "\n"),
		"reload_shell": c.ReloadShell,
	})
	err = ioutil.WriteFile(c.TomlDest, []byte(tomlStr), 0666)
	if err != nil {
		log.Warn("ioutil.WriteFile(%s, []byte(%s), 0666),err:%v", c.TomlDest, tomlStr, err)
	}
	err = conf.WriteFile(path.Join(DIR_TEMPLATES, c.TmplSrc), 0666, "#general by zgen")
	if err != nil {
		log.Warn("conf.WriteFile(%s, os.ModeType),err:%v", c.TmplSrc, err)
	}
	log.Info("new:%v", c)
	return err
}
