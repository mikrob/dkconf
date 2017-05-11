package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"text/template/parse"
	"unicode"
)

const (
	missingVarStr = "####### DKCONF : MISSING ENV VAR FOR GO TPL VALUE: %s, SHOULD BE %s #######"
)

var (
	sourceTplFile = flag.String("s", "", "absolute path to the source template file")
	targetFile    = flag.String("t", "", "absolute path to the target file generated")
	envPrefix     = flag.String("p", "APPCONF", "env var prefix")
)

type NginxConfig struct {
	Fqdn string
}

//ListTemplFields List field in templates
func ListTemplFields(t *template.Template) []string {
	return listNodeFields(t.Tree.Root, nil)
}

//listNodeFields list fields in templates nodes
func listNodeFields(node parse.Node, res []string) []string {
	if node.Type() == parse.NodeRange { // range list
		rangeNode := node.(*parse.RangeNode)
		listName := rangeNode.Pipe.Cmds[0].Args[0].String()
		formatedName := fmt.Sprintf("{{%s}}", listName)
		res = append(res, formatedName)
	}

	if node.Type() == parse.NodeIf {
		ifNode := node.(*parse.IfNode)
		ifVarName := ifNode.Pipe.Cmds[0].Args[0].String()
		formatedName := fmt.Sprintf("{{%s}}", ifVarName)
		res = append(res, formatedName)

	}

	if node.Type() == parse.NodeAction { // variable to interpret
		res = append(res, node.String())
	}

	if ln, ok := node.(*parse.ListNode); ok {
		for _, n := range ln.Nodes {
			res = listNodeFields(n, res)
		}
	}
	return res
}

//RemoveDuplicates remove duplicates string in an array of strings
func RemoveDuplicates(xs *[]string) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *xs {
		if !found[x] {
			found[x] = true
			(*xs)[j] = (*xs)[i]
			j++
		}
	}
	*xs = (*xs)[:j]
}

//initializeTemplate allow to initializeTemplate by creating template invocation and by listing field
func initializeTemplate() (*template.Template, error) {
	t, err := template.ParseFiles(*sourceTplFile)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return t, err
}

func SpaceMap(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

//extractFieldName list fields name in template
func extractFieldName(s string) string {
	re := regexp.MustCompile(`{{\s*\.(.+?)\s*}}`)
	match := re.FindStringSubmatch(s)
	return SpaceMap(match[1])
}

//formatEnvVar format an env
func formatEnvVar(value string) string {
	bashStyleField := replaceUpperWithUnderscore(value)
	return fmt.Sprintf("%s_%s", *envPrefix, strings.ToUpper(bashStyleField))
}

//replaceUpperWithUnderscore lookup at camelcase style words and split at each maj to allow an underscore insertion
func replaceUpperWithUnderscore(value string) string {
	var words []string
	l := 0
	for s := value; s != ""; s = s[l:] {
		l = strings.IndexFunc(s[1:], unicode.IsUpper) + 1
		if l <= 0 {
			l = len(s)
		}
		words = append(words, s[:l])
	}
	return strings.Join(words, "_")
}

//retrieveEnv list all field present in template and lookup at env var that match in bash style : A_B_C
func retrieveEnv(t *template.Template) (map[string]interface{}, []string) {
	fieldList := ListTemplFields(t)
	var missingList []string
	env := make(map[string]interface{})
	RemoveDuplicates(&fieldList)
	for _, field := range fieldList {
		realField := extractFieldName(field)
		formatedVar := formatEnvVar(realField)
		val, ok := os.LookupEnv(formatedVar)
		if ok {
			if strings.Contains(val, ",") { // list
				values := strings.Split(val, ",")
				env[realField] = values
			} else {
				if val == "true" || val == "false" { // boolean
					val, _ := strconv.ParseBool(val)
					env[realField] = val
				} else { // others
					env[realField] = val
				}
			}
		} else {
			env[realField] = fmt.Sprintf(missingVarStr, realField, formatedVar)
			//missingList = append(missingList, formatedVar)
		}
	}
	return env, missingList
}

//checkFileExists check if file exists in filesystem
func checkFileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

//parseTemplate parse the template with the given config map built in reading env var
func parseTemplate(t *template.Template, config map[string]interface{}) error {
	if *targetFile == "" { // if no target file is defined we output to stdout
		f := bufio.NewWriter(os.Stdout)
		defer f.Flush()
		err := t.Execute(f, config)
		if err != nil {
			log.Print("execute: ", err)
			return err
		}
	} else { // if we have target file we write to it
		f, err := os.Create(*targetFile)
		if err != nil {
			log.Println("create file: ", err)
			return err
		}
		err = t.Execute(f, config)
		if err != nil {
			log.Print("execute: ", err)
			return err
		}
		f.Close()
	}
	return nil
}

func main() {
	flag.Parse()

	if !checkFileExists(*sourceTplFile) {
		str := fmt.Sprintf("Source Template File does not exists : %s", *sourceTplFile)
		log.Println(str)
		os.Exit(1)
	}

	t, err := initializeTemplate()

	if err != nil {
		str := fmt.Sprintf("Cannot initialize template du to error : %s", err)
		log.Println(str)
		os.Exit(2)
	}

	env, _ := retrieveEnv(t)
	// if len(missings) != 0 {
	// 	fmt.Println("Some fields are missing in env : ", missings)
	// }

	parseTemplate(t, env)

}
