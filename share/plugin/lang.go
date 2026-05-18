package plugin

import "net/http"

var langMap = map[string]map[string]string{}

// RegisterLang 注册语言包，如 RegisterLang("zh", map[string]string{"user.not_found": "用户不存在"})
func RegisterLang(lang string, texts map[string]string) {
	if langMap[lang] == nil {
		langMap[lang] = texts
	} else {
		for k, v := range texts {
			langMap[lang][k] = v
		}
	}
}

// GetText 根据请求header中的lang获取翻译文本，未找到则返回key本身
func GetText(r *http.Request, key string) string {
	lang := r.Header.Get("lang")
	if lang == "" {
		lang = "zh"
	}
	if texts, ok := langMap[lang]; ok {
		if v, ok := texts[key]; ok {
			return v
		}
	}
	return key
}
