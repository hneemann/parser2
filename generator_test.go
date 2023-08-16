package parser2

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

type MyType interface {
	Name() string
}

type Simple struct {
}

func (s Simple) Name() string {
	return "simple"
}

func (s Simple) Matching(MyType) (MyType, error) {
	return Simple{}, nil
}

func (s Simple) NotMatching(float64) (MyType, error) {
	return Simple{}, nil
}

func (s Simple) NotMatching2(MyType) MyType {
	return Simple{}
}

func (s Simple) NotMatching3(MyType) (float64, error) {
	return 0, nil
}

type Pointer struct {
}

func (p *Pointer) Name() string {
	return "simple"
}

func (p *Pointer) Matching(MyType) (MyType, error) {
	return Simple{}, nil
}

func (p *Pointer) NotMatching(float64) (MyType, error) {
	return Simple{}, nil
}

func (p *Pointer) NotMatching2(MyType) MyType {
	return Simple{}
}

func (p *Pointer) NotMatching3(MyType) (float64, error) {
	return 0, nil
}

func Test_matchesInterface(t *testing.T) {
	tests := []struct {
		name       string
		value      MyType
		wantErr    string
		typeOf     reflect.Type
		methodName string
	}{
		{
			name:       "simple matching",
			value:      Simple{},
			typeOf:     reflect.TypeOf(Simple{}),
			methodName: "Matching",
			wantErr:    "",
		},
		{
			name:       "simple not matching",
			typeOf:     reflect.TypeOf(Simple{}),
			methodName: "NotMatching",
			wantErr:    "not match",
		},
		{
			name:       "simple not matching2",
			typeOf:     reflect.TypeOf(Simple{}),
			methodName: "NotMatching2",
			wantErr:    "wrong number",
		},
		{
			name:       "simple not matching3",
			typeOf:     reflect.TypeOf(Simple{}),
			methodName: "NotMatching3",
			wantErr:    "value needs to be assignable",
		},
		{
			name:       "pointer matching",
			value:      &Pointer{},
			typeOf:     reflect.TypeOf(&Pointer{}),
			methodName: "Matching",
			wantErr:    "",
		},
		{
			name:       "pointer not matching",
			typeOf:     reflect.TypeOf(&Pointer{}),
			methodName: "NotMatching",
			wantErr:    "not match",
		},
		{
			name:       "simple not matching2",
			typeOf:     reflect.TypeOf(&Pointer{}),
			methodName: "NotMatching2",
			wantErr:    "wrong number",
		},
		{
			name:       "simple not matching3",
			typeOf:     reflect.TypeOf(&Pointer{}),
			methodName: "NotMatching3",
			wantErr:    "value needs to be assignable",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			m, found := test.typeOf.MethodByName(test.methodName)
			assert.True(t, found)
			err := matches[MyType](m)
			if err != nil {
				// there was an error
				assert.True(t, test.wantErr != "", "no error expected")
				assert.True(t, strings.Contains(err.Error(), test.wantErr), "error has wrong message: "+err.Error())
			} else {
				// there was no error
				assert.True(t, test.wantErr == "", "expected error '"+test.wantErr+"'")

				f, err := methodByReflection(test.value, test.methodName)
				assert.NoError(t, err)
				_, err = f.Func([]MyType{test.value, Simple{}})
				assert.NoError(t, err)
			}
		})
	}
}

type MyFloat float64

func (f MyFloat) Matching(a MyFloat) (MyFloat, error) {
	return 5, nil
}

func (f MyFloat) NotMatching(a int) (MyFloat, error) {
	return 5, nil
}

func Test_matchesNoInterface(t *testing.T) {
	tests := []struct {
		name       string
		methodName string
		wantErr    string
		result     MyType
	}{
		{
			name:       "float matching",
			methodName: "Matching", //       must(reflect.TypeOf(MyFloat(0)).MethodByName("Matching")),
			wantErr:    "",
		},
		{
			name:       "float not matching",
			methodName: "NotMatching",
			wantErr:    "not match",
		},
	}
	for _, tt := range tests {
		test := tt
		t.Run(test.name, func(t *testing.T) {
			m, ok := reflect.TypeOf(MyFloat(0)).MethodByName(test.methodName)
			assert.True(t, ok)
			err := matches[MyFloat](m)
			if err != nil {
				// there was an error
				assert.True(t, test.wantErr != "", "no error expected")
				assert.True(t, strings.Contains(err.Error(), test.wantErr), "error has wrong message")
			} else {
				// there was no error
				assert.True(t, test.wantErr == "", "expected error '"+test.wantErr+"'")

				f, err := methodByReflection(MyFloat(0), test.methodName)
				assert.NoError(t, err)
				r, err := f.Func([]MyFloat{MyFloat(0), MyFloat(1)})
				assert.NoError(t, err)
				assert.EqualValues(t, 5, r)

			}
		})
	}
}

func TestPrintMatchingCode(t *testing.T) {
	PrintMatchingCode[MyType](Simple{})
	PrintMatchingCode[MyType](&Pointer{})
}
