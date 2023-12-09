package xmlWriter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	assert.Equal(t, "<map a=\"val\"/>\n", New().Open("map").Attr("a", "val").Close().String())
	assert.Equal(t, "<map a=\"val\">Test</map>\n", New().Open("map").Attr("a", "val").Write("Test").Close().String())
	assert.Equal(t, "<m>\n\t<e>Test</e>\n</m>\n", New().Open("m").Open("e").Write("Test").Close().Close().String())
	assert.Equal(t, "<map a=\"val\">test</map>\n", New().Open("map").Attr("a", "val").Write("test").Close().String())
	assert.Equal(t, "<a>\n\t<map a=\"val\"/>\n</a>\n", New().Open("a").Open("map").Attr("a", "val").Close().Close().String())
	assert.Equal(t, "<a>\n\t<b>\n\t\t<map a=\"val\"/>\n\t</b>\n</a>\n", New().Open("a").Open("b").Open("map").Attr("a", "val").Close().Close().Close().String())
	assert.Equal(t, "<map a=\"&lt;&amp;&gt;&apos;&quot;\"/>\n", New().Open("map").Attr("a", "<&>'\"").Close().String())
	assert.Equal(t, "<a>&lt;&amp;&gt;&apos;&quot;</a>\n", New().Open("a").Write("<&>'\"").Close().String())
	assert.Equal(t, "<a a=\"b\">\n\t<z>test</z>\n</a>\n", New().Open("a").Attr("a", "b").Open("z").Write("test").Close().Close().String())
	assert.Equal(t, "<map a=\"val\"></map>\n", New().AvoidShort().Open("map").Attr("a", "val").Close().String())
	assert.Equal(t, "<map></map>\n", New().AvoidShort().Open("map").Close().String())
}
