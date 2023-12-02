package value

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		exp string
		err string
	}{
		{"notFound(a)", "not found: notFound"},
		{"sin(1,2)", "number of args wrong"},
		{"{a:sin(1,2)}", "number of args wrong"},
	}

	fg := SetUpParser(New())
	for _, tt := range tests {
		test := tt
		t.Run(test.exp, func(t *testing.T) {
			_, err := fg.Generate(test.exp)
			if err == nil {
				t.Errorf("expected error %v", test.err)
			} else {
				assert.True(t, strings.Contains(err.Error(), test.err), "expected error to contain %v, got %v", test.err, err.Error())
			}
		})
	}
}
