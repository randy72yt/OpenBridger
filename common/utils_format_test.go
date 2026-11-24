package common

import (
	"testing"
	"html/template"

	"github.com/stretchr/testify/assert"
)

func TestBytes2Size(t *testing.T) {
	assert.Equal(t, "512 B", Bytes2Size(512))
	assert.Equal(t, "1 KB", Bytes2Size(1024))
	assert.Equal(t, "2 KB", Bytes2Size(2048))
	assert.Equal(t, "1 MB", Bytes2Size(1024*1024))
	assert.Equal(t, "1.00 GB", Bytes2Size(1024*1024*1024))
	assert.Equal(t, "1.50 GB", Bytes2Size(1024*1024*1024*3/2))
}

func TestSeconds2Time(t *testing.T) {
	assert.Equal(t, "0 秒", Seconds2Time(0))
	assert.Equal(t, "59 秒", Seconds2Time(59))
	assert.Equal(t, "1 分钟 5 秒", Seconds2Time(65))
	assert.Equal(t, "1 小时 1 分钟 1 秒", Seconds2Time(3661))
	assert.Equal(t, "1 天 0 秒", Seconds2Time(86400)) // 高位零单位跳过，末尾秒段恒显
}

func TestUnescapeHTMLAndMax(t *testing.T) {
	// 实际行为：仅做 template.HTML 类型转换（模板免转义），并不 unescape。
// 函数名与行为不符，属既有债务；锁定现状防意外变更。
	assert.Equal(t, template.HTML("a&lt;b&gt;&amp;c"), UnescapeHTML("a&lt;b&gt;&amp;c"))
	assert.Equal(t, 3, Max(1, 3))
	assert.Equal(t, -1, Max(-1, -2))
}

func TestInterface2StringAndBuildURL(t *testing.T) {
	assert.Equal(t, "42", Interface2String(42))
	assert.Equal(t, "text", Interface2String("text"))
	assert.Equal(t, "1.5", Interface2String(1.5))
	assert.Equal(t, "true", Interface2String(true))

	assert.Equal(t, "https://api.example.com/v1/chat", BuildURL("https://api.example.com", "/v1/chat"))
}
