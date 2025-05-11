package xmlWriter

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	assert.Equal(t, "<map a=\"val\"/>", New().Open("map").Attr("a", "val").Close().String())
	assert.Equal(t, "<map a=\"val\">Test</map>", New().Open("map").Attr("a", "val").Write("Test").Close().String())
	assert.Equal(t, "<m><e>Test</e></m>", New().Open("m").Open("e").Write("Test").Close().Close().String())
	assert.Equal(t, "<map a=\"val\">test</map>", New().Open("map").Attr("a", "val").Write("test").Close().String())
	assert.Equal(t, "<a><map a=\"val\"/></a>", New().Open("a").Open("map").Attr("a", "val").Close().Close().String())
	assert.Equal(t, "<a><b><map a=\"val\"/></b></a>", New().Open("a").Open("b").Open("map").Attr("a", "val").Close().Close().Close().String())
	assert.Equal(t, "<map a=\"&lt;&amp;&gt;&apos;&quot;\"/>", New().Open("map").Attr("a", "<&>'\"").Close().String())
	assert.Equal(t, "<a>&lt;&amp;&gt;&apos;&quot;</a>", New().Open("a").Write("<&>'\"").Close().String())
	assert.Equal(t, "<a a=\"b\"><z>test</z></a>", New().Open("a").Attr("a", "b").Open("z").Write("test").Close().Close().String())
	assert.Equal(t, "<map a=\"val\"></map>", New().AvoidShort().Open("map").Attr("a", "val").Close().String())
	assert.Equal(t, "<map></map>", New().AvoidShort().Open("map").Close().String())
}
