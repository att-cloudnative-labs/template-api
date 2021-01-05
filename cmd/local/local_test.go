package main

import (
	"os"
	"testing"
)

func Test_folderWithSeparator(t *testing.T) {
	const sep = string(os.PathSeparator)
	type testCase struct {
		Input    string
		Expected string
	}
	cases := []testCase{
		{"myfolder", "myfolder" + sep},
		{"myfolder" + sep, "myfolder" + sep},
		{"myfolder" + sep + "sub", "myfolder" + sep + "sub" + sep},
		{"myfolder" + sep + "sub" + sep, "myfolder" + sep + "sub" + sep},
		{"", ""},
	}
	for _, testCase := range cases {
		value := folderWithSeparator(testCase.Input)
		if value != testCase.Expected {
			t.Errorf("Expected '%s' to equal '%s'", value, testCase.Expected)
		}
	}
}
