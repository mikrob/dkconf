package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"text/template"
)

var (
	varStandard           = "this_is_a_config_value"
	varList               = "ab,cd,ef,gh,ij"
	varBool               = "true"
	varNotExists          = fmt.Sprintf(missingVarStr, "VarNotExists", "APPCONF_VAR_NOT_EXISTS")
	testTemplate          = "{{.VarStandard}} items are made of {{.VarList}} are you ok ? {{.VarBool}} And ... {{.VarNotExists}}"
	testTemplateBadSyntax = "{{.VarStandard} items are made of {.VarList}} are you ok ? {{.VarBool}} And ... {{.VarNotExists}}"
	parsedTemplate        = "this_is_a_config_value items are made of [ab cd ef gh ij] are you ok ? true And ... ####### DKCONF : MISSING ENV VAR FOR GO TPL VALUE: VarNotExists, SHOULD BE APPCONF_VAR_NOT_EXISTS #######"
)

func TestRemoveDuplicates(t *testing.T) {
	list := []string{"truc", "bidule", "truc", "machin", "truc", "machin"}
	RemoveDuplicates(&list)

	wantedList := []string{"truc", "bidule", "machin"}
	if !reflect.DeepEqual(list, wantedList) {
		t.Errorf("Still have duplicate in list : %v, %v", list, wantedList)
	}
}

func TestSpaceMap(t *testing.T) {
	strWithSpaces := " .Bidule "
	strChomped := SpaceMap(strWithSpaces)
	if strChomped != ".Bidule" {
		t.Error("Still have space in result of SpaceMap")
	}
}

func TestExtractFieldName(t *testing.T) {
	str := "{{.Bidule }}"
	fieldName := extractFieldName(str)
	if fieldName != "Bidule" {
		t.Errorf("extractFieldName should have returned [Bidule] and not [%s]", fieldName)
	}

}

func TestFormatEnvVar(t *testing.T) {
	strToFormat := "MyValueIsCamelCase"
	strFormated := formatEnvVar(strToFormat)

	if strFormated != "APPCONF_MY_VALUE_IS_CAMEL_CASE" {
		t.Errorf("Result should be MY_VALUE_IS_CAMEL_CASE, got : %s", strFormated)
	}
}

func MakeConfig() map[string]interface{} {
	goVarList := strings.Split(varList, ",")
	wantedMap := make(map[string]interface{})
	wantedMap["VarStandard"] = varStandard
	wantedMap["VarList"] = goVarList
	wantedMap["VarBool"] = true
	wantedMap["VarNotExists"] = varNotExists
	return wantedMap
}

func TestRetrieveEnv(t *testing.T) {

	os.Setenv("APPCONF_VAR_STANDARD", varStandard)
	os.Setenv("APPCONF_VAR_LIST", varList)
	os.Setenv("APPCONF_VAR_BOOL", varBool)

	tmpl, _ := prepareTemplate(template.New("test")).Parse(testTemplate)
	config, _ := retrieveEnv(tmpl)

	wantedMap := MakeConfig()

	if !reflect.DeepEqual(config, wantedMap) {
		t.Errorf("Map are not equal want : %v, got : %v", wantedMap, config)
	}

}

func TestCheckFileExists(t *testing.T) {
	file, _ := ioutil.TempFile(os.TempDir(), "prefix")
	defer os.Remove(file.Name())

	if !checkFileExists(file.Name()) {
		t.Errorf("File : %s/%s exists so function is buggy", os.TempDir(), file.Name())
	}
}

type ParseFunc func(t *template.Template, config map[string]interface{}) error

func CaptureStdOut(function ParseFunc, t2 *template.Template, config2 map[string]interface{}) string {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	function(t2, config2)
	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout
	return string(out)
}

func TestInitializeTemplate(t *testing.T) {
	f, _ := os.Create("/tmp/tpl.tpl")
	w := bufio.NewWriter(f)
	w.WriteString(testTemplate)
	*sourceTplFile = "/tmp/tpl.tpl"
	tpl, err := initializeTemplate()
	if tpl == nil {
		t.Error("Template should not be nil")
	}
	if err != nil {
		t.Error("Err should be nil")
	}

}

func TestParseTemplate(t *testing.T) {
	tmpl, _ := prepareTemplate(template.New("test")).Parse(testTemplate)
	config := MakeConfig()
	stdout := CaptureStdOut(parseTemplate, tmpl, config)

	if stdout != parsedTemplate {
		t.Errorf("Generated template is not that what is waited, got : %s", stdout)
	}
}

func TestParseTemplateWithBadSyntax(t *testing.T) {
	_, errTpl := prepareTemplate(template.New("test")).Parse(testTemplateBadSyntax)
	if errTpl == nil {
		t.Error("Error in template instanciation")
	}
}

func TestParseTemplateIsIterable(t *testing.T) {
	assertParsed(t, "{{ if is_iterable .VarList}}Yes, it is{{ else }}No, it is not{{ end }}", "Yes, it is")
}

func TestParseTemplateWithIsIterableAndIterableVariable(t *testing.T) {
	assertParsed(t, "{{ if is_iterable .VarStandard}}Yes, it is{{ else }}No, it is not{{ end }}", "No, it is not")
}

func TestParseTemplateWithUpper(t *testing.T) {
	assertParsed(t, "{{ \"hello\" | upper }}", "HELLO")
	assertParsed(t, "{{ \"hELLO\" | upper }}", "HELLO")
	assertParsed(t, "{{ \"HELLO\" | upper }}", "HELLO")
	assertParsed(t, "{{ \"hello 1\" | upper }}", "HELLO 1")
}

func TestParseTemplateWithTrim(t *testing.T) {
	assertParsed(t, "{{ \" \" | trim }}", "")
	assertParsed(t, "{{ \" abc \" | trim }}", "abc")
	assertParsed(t, "{{ \" abc\" | trim }}", "abc")
	assertParsed(t, "{{ \"abc \" | trim }}", "abc")
}

func TestParseTemplateWithLower(t *testing.T) {
	assertParsed(t, "{{ \"HellO\" | lower }}", "hello")
	assertParsed(t, "{{ \"hELLo\" | lower }}", "hello")
	assertParsed(t, "{{ \"hello\" | lower }}", "hello")
	assertParsed(t, "{{ \"HELLO 1\" | lower }}", "hello 1")
}

func TestParseTemplateWithCombinedUpperAndLower(t *testing.T) {
	assertParsed(t, "{{ \"HellO\" | lower | upper }}", "HELLO")
}

func TestParseTemplateWithReplace(t *testing.T) {
	assertParsed(t, "{{ replace \"abcde ab tab\" \"ab\" \"xy\" }}", "xycde xy txy")
}

func TestParseTemplateWithUcwords(t *testing.T) {
	assertParsed(t, "{{ \"wORD1 woRd2 Word3 word4 worD5 WORD6\" | ucwords }}", "WORD1 WoRd2 Word3 Word4 WorD5 WORD6")
}

func TestParseTemplateWithDefaultAndNonEmptyValue(t *testing.T) {
	assertParsed(t, "{{ default \"hello\" \"Hello World ;-)\" }}", "hello")
}

func TestParseTemplateWithDefaultAndEmptyValue(t *testing.T) {
	assertParsed(t, "{{ default \"\" \"Hello World ;-)\" }}", "Hello World ;-)")
}

func TestParseTemplateWithDefaultAndMissingEnv(t *testing.T) {
	assertParsed(t, "{{ default .MyMissingEnvVar \"Hello World ;-)\" }}", "Hello World ;-)")
}

func assertParsed(t *testing.T, tpl string, parsed string) {
	tmpl, _ := prepareTemplate(template.New("test")).Parse(tpl)
	config := MakeConfig()
	stdout := CaptureStdOut(parseTemplate, tmpl, config)

	if stdout != parsed {
		t.Errorf("Generated template is not that what is waited, got : %s", stdout)
	}
}