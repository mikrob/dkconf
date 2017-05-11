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

	tmpl, _ := template.New("test").Parse(testTemplate)
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

func TestParseTemplate(t *testing.T) {
	tmpl, _ := template.New("test").Parse(testTemplate)
	config := MakeConfig()
	stdout := CaptureStdOut(parseTemplate, tmpl, config)

	if stdout != parsedTemplate {
		t.Errorf("Generated template is not that what is waited, got : %s", stdout)
	}
}

func TestParseTemplateWithBadSyntax(t *testing.T) {
	_, errTpl := template.New("test").Parse(testTemplateBadSyntax)
	if errTpl == nil {
		t.Error("Error in template instanciation")
	}
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
