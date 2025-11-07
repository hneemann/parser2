package value

import (
	"bytes"
	"errors"
	"github.com/hneemann/parser2/funcGen"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

type String string

func (s String) ToList() (*List, bool) {
	return nil, false
}

func (s String) ToMap() (Map, bool) {
	return EmptyMap, false
}

func (s String) ToFloat() (float64, bool) {
	return 0, false
}

func (s String) ToString(funcGen.Stack[Value]) (string, error) {
	return string(s), nil
}

func (s String) String() string {
	return "\"" + string(s) + "\""
}

func (s String) Contains(st funcGen.Stack[Value]) (Value, error) {
	if s2, ok := st.Get(1).(String); ok {
		return Bool(strings.Contains(string(s), string(s2))), nil
	} else {
		return nil, errors.New("contains needs a string as argument")
	}
}

func (s String) IndexOf(st funcGen.Stack[Value]) (Value, error) {
	if s2, ok := st.Get(1).(String); ok {
		return Int(strings.Index(string(s), string(s2))), nil
	} else {
		return nil, errors.New("indexOf needs a string as argument")
	}
}

func (s String) Split(st funcGen.Stack[Value]) (Value, error) {
	if s2, ok := st.Get(1).(String); ok {
		return NewListConvert(func(s string) (Value, error) { return String(s), nil }, strings.Split(string(s), string(s2))), nil
	} else {
		return nil, errors.New("split needs a string as argument")
	}
}

func (s String) Cut(st funcGen.Stack[Value]) (Value, error) {
	if p, ok := st.Get(1).(Int); ok {
		if n, ok := st.Get(2).(Int); ok {
			str := string(s)
			for i := 0; i < int(p); i++ {
				_, l := utf8.DecodeRuneInString(str)
				str = str[l:]
				if len(str) == 0 {
					return String(""), nil
				}
			}
			var res bytes.Buffer
			if n <= 0 {
				n = math.MaxInt
			}
			for i := 0; i < int(n); i++ {
				r, l := utf8.DecodeRuneInString(str)
				res.WriteRune(r)
				str = str[l:]
				if len(str) == 0 {
					return String(res.String()), nil
				}
			}
			return String(res.String()), nil
		}
	}
	return nil, errors.New("cut requires integers as arguments (pos,len)")
}

func (s String) Behind(st funcGen.Stack[Value]) (Value, error) {
	l := strings.Split(string(s), "\n")
	if pre, ok := st.Get(1).(String); ok {
		for _, e := range l {
			p := strings.Index(e, string(pre))
			if p >= 0 {
				r := e[p+len(pre):]
				return String(strings.TrimSpace(r)), nil
			}
		}
		return String(""), nil
	} else {
		return nil, errors.New("behind needs a string as argument")
	}
}

func (s String) BehindList(st funcGen.Stack[Value]) (Value, error) {
	var foundItems []string
	lines := strings.Split(string(s), "\n")
	if kl, ok := st.Get(1).(String); ok {
		keyLine := strings.TrimSpace(string(kl))
		found := false
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if found {
				if line == "" {
					break
				} else {
					foundItems = append(foundItems, line)
				}
			} else {
				if line == keyLine {
					found = true
				}
			}
		}
		return NewListConvert(func(s string) (Value, error) { return String(s), nil }, foundItems), nil
	} else {
		return nil, errors.New("behind needs a string as argument")
	}
}

func (s String) Replace(st funcGen.Stack[Value]) (Value, error) {
	if oldStr, ok := st.Get(1).(String); ok {
		if newStr, ok := st.Get(2).(String); ok {
			return String(strings.Replace(string(s), string(oldStr), string(newStr), -1)), nil
		}
	}
	return nil, errors.New("replace needs two strings (old,new) as arguments")
}

func (s String) ParseToFloat() (Value, error) {
	f, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		return nil, err
	}
	return Float(f), nil
}

func (s String) ParseToInt() (Value, error) {
	i, err := strconv.Atoi(string(s))
	if err != nil {
		return nil, err
	}
	return Int(i), nil
}

func createStringMethods() MethodMap {
	return MethodMap{
		"len": MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) { return Int(len(string(str))), nil }).
			SetMethodDescription("Returns the length of the string."),
		"string": MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str, nil }).
			SetMethodDescription("Returns the string itself."),
		"trim": MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) {
			return String(strings.TrimSpace(string(str))), nil
		}).SetMethodDescription("Returns the string without leading and trailing spaces."),
		"toLower": MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) {
			return String(strings.ToLower(string(str))), nil
		}).SetMethodDescription("Returns the string in lower case."),
		"toUpper": MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) {
			return String(strings.ToUpper(string(str))), nil
		}).SetMethodDescription("Returns the string in upper case."),
		"contains": MethodAtType(1, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.Contains(stack) }).
			SetMethodDescription("substr",
				"Returns true if the string contains the substr."),
		"indexOf": MethodAtType(1, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.IndexOf(stack) }).
			SetMethodDescription("substr",
				"Returns the index of the first occurrence of substr in the string. Returns -1 if not found."),
		"split": MethodAtType(1, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.Split(stack) }).
			SetMethodDescription("sep",
				"Splits the string at the separator and returns a list of strings."),
		"cut": MethodAtType(2, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.Cut(stack) }).
			SetMethodDescription("pos", "len",
				"Returns a substring starting at pos with length len. "+
					"If len is negative, the rest of the string is returned."),
		"behind": MethodAtType(1, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.Behind(stack) }).
			SetMethodDescription("prefix", "Returns the string behind the prefix up to the next newline."),
		"behindList": MethodAtType(1, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.BehindList(stack) }).
			SetMethodDescription("header", "Returns the lines following behind the header line up to the next empty line."),
		"replace": MethodAtType(2, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.Replace(stack) }).
			SetMethodDescription("old", "new", "Replaces all occurrences of old with new."),
		"toFloat": MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.ParseToFloat() }).
			SetMethodDescription("Parses the string to a float."),
		"toInt": MethodAtType(0, func(str String, stack funcGen.Stack[Value]) (Value, error) { return str.ParseToInt() }).
			SetMethodDescription("Parses the string to an int."),
	}
}

func (s String) GetType() Type {
	return StringTypeId
}
