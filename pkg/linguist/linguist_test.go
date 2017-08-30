package linguist

import (
	"testing"
)

const appPythonPath = "testdata/app-python"

const appEmptydirPath = "testdata/app-emptydir"

func TestProcessDir(t *testing.T) {
	output, err := ProcessDir(appPythonPath)
	if err != nil {
		t.Error("expected detect to pass")
	}
	if output[0].Language != "Python" {
		t.Errorf("expected output == 'Python', got '%s'", output[0].Language)
	}

	// test with a bad dir
	if _, err := ProcessDir("/dir/does/not/exist"); err == nil {
		t.Error("expected err when running detect with a dir that does not exist")
	}

	// test an application that should fail detection
	output, _ = ProcessDir(appEmptydirPath)
	if len(output) != 0 {
		t.Errorf("expected no languages detected, got '%d'", len(output))
	}
}

func TestGetAlias(t *testing.T) {
	testcases := map[string]string{
		"maven pom": "Java",
		"mAvEn POM": "Java",
		"c#":        "csharp",
		"Python":    "Python",
	}

	for packName, expectedAlias := range testcases {
		alias := Alias(&Language{Language: packName})
		if alias.Language != expectedAlias {
			t.Errorf("Expected alias to be '%s', got '%s'", expectedAlias, alias.Language)
		}
	}
}
