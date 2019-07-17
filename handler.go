package gdoc

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/techoner/gdoc/resources/assets"
	"github.com/techoner/gdoc/resources/source"
	"github.com/techoner/gdoc/resources/views"
	"gopkg.in/russross/blackfriday.v1"
	"gopkg.in/yaml.v2"
)

var DefaultHandler = &Handler{
	DefaultVersionName: "default",
	DocsDir:            "storage/docs",
	PrefixUri:          "docs",
}

type Handler struct {
	DefaultVersionName string
	DocsDir            string
	PrefixUri          string
}

func Handle(name string) []byte {
	return DefaultHandler.Handle(name)
}

func (h *Handler) Handle(name string) []byte {

	version := h.DefaultVersionName
	baseName := "index.html"
	dirName := ""
	versions := make(map[string]string)

	if len(name) > 0 {
		bname := path.Base(name)
		fext := path.Ext(name)
		fname := strings.TrimSuffix(bname, fext)
		dname := path.Dir(name)
		if dname == "." {
			dname = "/"
		}

		fragment := strings.SplitN(strings.Trim(dname, "/"), "/", 2)

		v := name
		if len(fragment) > 0 {
			v = fragment[0]
		}

		versions = h.GetVersion(v)
		if len(versions) > 0 {
			version = v
		}

		dirName = strings.Replace(dname, version, "", -1)

		if version != fname && bname != "" && bname != "/" {
			baseName = bname
		}
	}

	versions = h.GetVersion(version)
	sidebar := h.GetSidebar(version)

	contentFileName := path.Join(dirName, baseName)
	content := h.GetContent(version, contentFileName)

	var buf bytes.Buffer
	t := template.New("")
	tmpl, err := t.Parse(views.Index())
	if err != nil {
		panic(err)
	}
	currentVersionTitle := ""
	if len(version) == 0 {
		currentVersionTitle = versions[h.DefaultVersionName]
	} else {
		currentVersionTitle = versions[version]
	}

	if version == h.DefaultVersionName {
		version = ""
	}

	tmpl.ExecuteTemplate(&buf, "gloc", map[string]interface{}{
		"css":                   assets.Index,
		"sidebar":               template.JS(sidebar),
		"content":               template.HTML(content),
		"versions":              versions,
		"current_version":       version,
		"current_version_title": currentVersionTitle,
		"prefix_uri":            path.Join("/", h.PrefixUri) + "/",
		"basePath":              path.Join("/", h.PrefixUri, version) + "/",
		"contentFileName":       strings.TrimLeft(contentFileName, "/"),
		"default_version_name":  h.DefaultVersionName,
	})

	return buf.Bytes()
}

func (h *Handler) GetVersion(version string) map[string]string {
	versions := make(map[string]string)

	p := h.GetStorageFilePath("versions.yml")

	if !isFile(p) {
		return versions
	}

	versions = yamlParseFile(p)
	_, ok := versions[version]
	if !ok {
		return nil
	}

	_, ok = versions["default"]
	if !ok {
		versions["default"] = "默认版本"
	}

	return versions
}

func (h *Handler) GetContent(version string, p string) string {
	if version == h.DefaultVersionName {
		version = ""
	}
	p = h.GetStorageFilePath(
		path.Join(version, "_source", strings.TrimSuffix(p, path.Ext(p))),
	)

	p = p + ".md"

	exist := isFile(p)
	var content []byte
	if !exist {
		content = []byte(source.Default)
	} else {
		content, _ = ioutil.ReadFile(p)
	}
	commonHtmlFlags := 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_DASHES |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	renderer := blackfriday.HtmlRenderer(commonHtmlFlags, "", "")
	extensions := 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_HEADER_IDS |
		blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
		blackfriday.EXTENSION_DEFINITION_LISTS |
		blackfriday.EXTENSION_AUTO_HEADER_IDS
	content = blackfriday.Markdown(content, renderer, extensions)

	return string(content)
}

func (h *Handler) ParseSidebar(version string) map[string]map[string]string {
	sidebars := make(map[string]map[string]string)
	p := "sidebar.yml"
	if version != h.DefaultVersionName {
		p = version + "/" + p
	}

	p = h.GetStorageFilePath(p)

	if !isFile(p) {
		return sidebars
	}

	data, _ := ioutil.ReadFile(p)

	yaml.Unmarshal(data, &sidebars)

	return sidebars
}

func (h *Handler) GetSidebar(version string) string {
	p := "sidebar.yml"
	if version != h.DefaultVersionName {
		p = version + "/" + p
	}

	p = h.GetStorageFilePath(p)

	if !isFile(p) {
		return ""
	}

	data, _ := ioutil.ReadFile(p)

	return string(data)
}

func (h *Handler) GetStorageFilePath(name string) string {
	return path.Join(h.DocsDir, name)
}

func isFile(p string) bool {

	f, err := os.Stat(p)
	if err != nil {
		return false
	}

	return !f.IsDir()
}

func yamlParseFile(p string) map[string]string {
	versions := make(map[string]string)

	data, _ := ioutil.ReadFile(p)
	yaml.Unmarshal(data, &versions)

	return versions
}
