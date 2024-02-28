package templates

import (
	_ "embed"
	"html/template"
	"sync"
)

var (
	//go:embed mail/layout.html
	MailTemplate string
)

var templateCache = make(map[string]*template.Template)
var templateCacheMux = &sync.RWMutex{}

func GetMailTemplate() *template.Template {
	return getTemplate("mail", func() *template.Template {
		return template.Must(template.New("mail").Parse(MailTemplate))
	})
}

func getTemplate(name string, create func() *template.Template) *template.Template {
	templateCacheMux.RLock()
	if templateCache[name] != nil {
		defer templateCacheMux.RUnlock()
		return templateCache[name]
	}
	templateCacheMux.RUnlock()

	tmpl := create()
	templateCacheMux.Lock()
	defer templateCacheMux.Unlock()
	templateCache[name] = tmpl
	return templateCache[name]
}
