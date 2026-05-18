package locales

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/fs185085781/v9os/internal/ioc/uioc"
	"github.com/gin-gonic/gin"
)

//go:embed *.json
var localeFS embed.FS

// 用于同时给前端和后端使用
func GetModelLang(lang string) map[string]interface{} {
	return getCacheJSON(fmt.Sprintf("model-%s.json", lang))
}

func flattenMap(prefix string, nested map[string]interface{}) map[string]string {
	flat := make(map[string]string)
	for k, v := range nested {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}
		switch val := v.(type) {
		case map[string]interface{}:
			for ik, iv := range flattenMap(fullKey, val) {
				flat[ik] = iv
			}
		case string:
			flat[fullKey] = val
		default:
		}
	}
	return flat
}

// 修改现有方法
func getCacheTextMap(rPath string) map[string]string {
	c := uioc.Cache()
	var res map[string]string
	has, _ := c.GetObjectRetry("locale:text:"+rPath, &res)
	if !has {
		res = flattenMap("", getCacheJSON(rPath))
		if res != nil {
			c.SetObjectRetry("locale:text:"+rPath, res, 30*time.Minute)
		} else {
			return make(map[string]string)
		}
	}
	return res
}
func getCacheJSON(rPath string) map[string]interface{} {
	c := uioc.Cache()
	var res map[string]interface{}
	has, _ := c.GetObjectRetry("locale:json:"+rPath, &res)
	if !has {
		res = getJSON(rPath)
		if res != nil {
			c.SetObjectRetry("locale:json:"+rPath, res, 1*time.Minute)
		} else {
			return make(map[string]interface{})
		}
	}
	return res
}
func getJSON(rPath string) map[string]interface{} {
	data, err := localeFS.ReadFile(rPath)
	if err != nil {
		return nil
	}
	var res map[string]interface{}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil
	}
	return res
}

func GetText(lang string, key string) string {
	var flat map[string]string
	if strings.HasPrefix(key, "model.") {
		flat = getCacheTextMap(fmt.Sprintf("model-%s.json", lang))
	} else if strings.HasPrefix(key, "common.") {
		flat = getCacheTextMap(fmt.Sprintf("common-%s.json", lang))
	} else {
		return key
	}
	v := flat[key]
	if v == "" {
		return key
	}
	return v
}

func GetLang(ctx *gin.Context) string {
	lang := ctx.Query("lang")
	if lang == "" {
		lang = ctx.GetHeader("lang")
		if lang == "" {
			lang = "zh"
		}
	}
	return lang
}

func GetTextCtx(c *gin.Context, key string) string {
	return GetText(GetLang(c), key)
}
