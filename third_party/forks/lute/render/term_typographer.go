// Copyright (c) 2019-present, Scribli


package render

import (
	"bytes"
	"unicode/utf8"

	"github.com/icha-senpai/note/third_party/forks/lute/lex"
	"github.com/icha-senpai/note/third_party/forks/lute/util"
)

func (r *BaseRenderer) FixTermTypo(tokens []byte) []byte {
	return r.fixTermTypo0(tokens)
}

func (r *BaseRenderer) fixTermTypo0(tokens []byte) []byte {
	length := len(tokens)
	var token byte
	var i, j, k, l int
	var before, after byte
	var originalTerm []byte
	for ; i < length; i++ {
		token = tokens[i]
		if isNotTerm(token) {
			continue
		}
		if 1 <= i {
			before = tokens[i-1]
			if !isNotTerm(before) {
				continue
			}
		}
		if lex.IsASCIIPunct(before) {
			continue
		}

		for j = i; j < length; j++ {
			after = tokens[j]
			if isNotTerm(after) || lex.ItemDot == after {
				break
			}
		}
		if lex.IsASCIIPunct(after) {
			continue
		}

		originalTerm = bytes.ToLower(tokens[i:j])
		if to, ok := r.Options.Terms[util.BytesToStr(originalTerm)]; ok {
			l = 0
			for k = i; k < j; k++ {
				tokens[k] = to[l]
				l++
			}
		}
	}

	return tokens
}

func isNotTerm(token byte) bool {
	return token >= utf8.RuneSelf || lex.IsWhitespace(token) || lex.IsASCIIPunct(token)
}

func NewTerms() (ret map[string]string) {
	ret = make(map[string]string, len(terms))
	for k, v := range terms {
		ret[k] = v
	}
	return
}

var terms = map[string]string{
	"flutter":       "Flutter",
	"netty":         "Netty",
	"jetty":         "Jetty",
	"tomcat":        "Tomcat",
	"jdbc":          "JDBC",
	"mariadb":       "MariaDB",
	"ipfs":          "IPFS",
	"saas":          "SaaS",
	"paas":          "PaaS",
	"iaas":          "IaaS",
	"ioc":           "IoC",
	"freemarker":    "FreeMarker",
	"ruby":          "Ruby",
	"rails":         "Rails",
	"mina":          "Mina",
	"puppet":        "Puppet",
	"vagrant":       "Vagrant",
	"chef":          "Chef",
	"beego":         "Beego",
	"gin":           "Gin",
	"iris":          "Iris",
	"php":           "PHP",
	"ssh":           "SSH",
	"web":           "Web",
	"websocket":     "WebSocket",
	"api":           "API",
	"css":           "CSS",
	"html":          "HTML",
	"json":          "JSON",
	"jsonp":         "JSONP",
	"xml":           "XML",
	"yaml":          "YAML",
	"csv":           "CSV",
	"soap":          "SOAP",
	"ajax":          "AJAX",
	"messagepack":   "MessagePack",
	"javascript":    "JavaScript",
	"java":          "Java",
	"jsp":           "JSP",
	"restful":       "RESTFul",
	"graphql":       "GraphQL",
	"gorm":          "GORM",
	"orm":           "ORM",
	"oauth":         "OAuth",
	"facebook":      "Facebook",
	"github":        "GitHub",
	"gist":          "Gist",
	"heroku":        "Heroku",
	"twitter":       "Twitter",
	"youtube":       "YouTube",
	"dynamodb":      "DynamoDB",
	"mysql":         "MySQL",
	"postgresql":    "PostgreSQL",
	"sqlite":        "SQLite",
	"memcached":     "Memcached",
	"mongodb":       "MongoDB",
	"redis":         "Redis",
	"elasticsearch": "Elasticsearch",
	"solr":          "Solr",
	"Scribli":       "Scribli",
	"sphinx":        "Sphinx",
	"linux":         "Linux",
	"ubuntu":        "Ubuntu",
	"centos":        "CentOS",
	"centos7":       "CentOS7",
	"redhat":        "RedHat",
	"gitlab":        "GitLab",
	"jquery":        "jQuery",
	"angularjs":     "AngularJS",
	"ffmpeg":        "FFmpeg",
	"git":           "Git",
	"svn":           "SVN",
	"vim":           "VIM",
	"emacs":         "Emacs",
	"sublime":       "Sublime",
	"virtualbox":    "VirtualBox",
	"safari":        "Safari",
	"chrome":        "Chrome",
	"ie":            "IE",
	"firefox":       "Firefox",
	"iterm":         "iTerm",
	"iterm2":        "iTerm2",
	"iwork":         "iWork",
	"itunes":        "iTunes",
	"iphoto":        "iPhoto",
	"ibook":         "iBook",
	"imessage":      "iMessage",
	"photoshop":     "Photoshop",
	"excel":         "Excel",
	"powerpoint":    "PowerPoint",
	"ios":           "iOS",
	"iphone":        "iPhone",
	"ipad":          "iPad",
	"android":       "Android",
	"imac":          "iMac",
	"macbook":       "MacBook",
	"vps":           "VPS",
	"vpn":           "VPN",
	"cpu":           "CPU",
	"spring":        "Spring",
	"springboot":    "SpringBoot",
	"springcloud":   "SpringCloud",
	"springmvc":     "SpringMVC",
	"mybatis":       "MyBatis",
	"qq":            "QQ",
	"sql":           "SQL",
	"markdown":      "Markdown",
	"jdk":           "JDK",
	"openjdk":       "OpenJDK",
	"cors":          "CORS",
	"protobuf":      "Protobuf",
	"google":        "Google",
	"ibm":           "IBM",
	"oracle":        "Oracle",
	"typora":        "Typora",
}
