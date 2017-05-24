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
	"reflect"
	"path/filepath"
)

const (
	missingVarStr = "####### DKCONF : MISSING ENV VAR FOR GO TPL VALUE: %s, SHOULD BE %s #######"
)

var (
	sourceTplFile = flag.String("s", "", "absolute path to the source template file")
	targetFile    = flag.String("t", "", "absolute path to the target file generated")
	envPrefix     = flag.String("p", "APPCONF", "env var prefix")
	envList map[string]interface{} = nil
	globalEnvList map[string]interface{} = nil
)

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
	var t *template.Template = template.New(filepath.Base(*sourceTplFile))
	var err error
	prepareTemplate(t)
	t, err = t.ParseFiles(*sourceTplFile)
	if err != nil {
		log.Print(err)
		return nil, err
	}
	return t, err
}

func prepareTemplate(t * template.Template) (* template.Template) {
	t.Funcs(template.FuncMap{
		"is_iterable": func(v interface{}) bool {
			vr := reflect.ValueOf(v)
			switch vr.Kind() {
			case reflect.String, reflect.Invalid, reflect.Bool:
				return false
			case reflect.Slice, reflect.Array, reflect.Map:
				return true
			default:
				return false
			}
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"ucwords": strings.Title,
		"title": strings.Title,
		"trim": strings.TrimSpace,
		"slugify": slugify,
		"underscore": underscore,
		"snakize": snakize,
		"envname": envname,
		"env": envvalue,
		"global_env": globalenvvalue,
		"split": func (optional_params ...string) []string {
			var v, sep string
			if (len(optional_params) >= 2) {
				v = optional_params[0]
				sep = optional_params[1]
			} else if (len(optional_params) >= 1) {
				v = optional_params[0]
				sep = " "
			}

			return strings.Split(v, sep)
		},
		"concat": func (optional_params ...string) string {
			return strings.Join(optional_params,"")
		},
		"sprintf": fmt.Sprintf,
		"contains": strings.Contains,
		"comma_split": func (v string) []string {
			var sep string = ","

			return strings.Split(v, sep)
		},
		"match": func (v string, r string) bool {
			var re = regexp.MustCompile(r)
			return re.MatchString(v)
		},
		"regexp_replace": func (v string, r string, new string) string {
			var re = regexp.MustCompile(r)
			return re.ReplaceAllString(v, new)
		},
		"env_list": func (v string) []string {
			var sep string = ","

			return strings.Split(envvalue(v).(string), sep)
		},
		"dump": func (v interface{}) string {
			return fmt.Sprintf("%+v", v)
		},
		"is_empty": func (v interface{}) bool {
			vr := reflect.ValueOf(v)
			switch vr.Kind() {
			case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
				return 0 >= vr.Len()
			case reflect.Invalid:
				return true
			case reflect.Bool:
				return !v.(bool)
			default:
				return false
			}
		},
		"is_not_empty": func (v interface{}) bool {
			vr := reflect.ValueOf(v)
			switch vr.Kind() {
			case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
				return 0 < vr.Len()
			case reflect.Invalid:
				return false
			case reflect.Bool:
				return v.(bool)
			default:
				return true
			}
		},
		"join": func (v interface{}, sep string) string {
			vr := reflect.ValueOf(v)
			switch vr.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
				return strings.Join(v.([]string), sep)
			case reflect.String:
				return v.(string)
			case reflect.Invalid, reflect.Bool:
				return ""
			default:
				return ""
			}
		},
		"is_enabled": func (v interface{}) bool {
			vr := reflect.ValueOf(v)
			switch vr.Kind() {
			case reflect.Slice, reflect.Array, reflect.Map:
				return 0 < vr.Len()
			case reflect.String:
				return 0 < vr.Len() && "0" != v.(string)
			case reflect.Invalid:
				return false
			case  reflect.Bool:
				return v.(bool)
			case  reflect.Int:
				return 0 != v.(int)
			default:
				return true
			}
		},
		"replace": func(s string, old string, new string) string {
			return strings.Replace(s, old, new, -1)
		},
		"default": func (v interface{}, d interface{}) interface{} {
			vr := reflect.ValueOf(v)
			switch vr.Kind() {
			case reflect.Invalid:
				return d
			case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
				if vr.Len() == 0 {
					return d
				}
				return v
			case reflect.Bool:
				if !vr.Bool() {
					return d
				}
				return v
			default:
				return v
			}
		},
		"length": func (v interface{}) int {
			vr := reflect.ValueOf(v)
			switch vr.Kind() {
			case reflect.Invalid:
				return 0
			case reflect.String, reflect.Slice, reflect.Array, reflect.Map:
				return vr.Len();
			case reflect.Bool:
				return 0
			default:
				return 0
			}
		},
	})
	return t
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
	re := regexp.MustCompile(`\{\{\s*[^\.]*\.([^\."]+).*\s*\}\}`)
	match := re.FindStringSubmatch(s)
	if (0 < len(match)) {
		return SpaceMap(match[1])
	}
	return ""
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

func envs() map[string]interface{} {
	if (nil == envList) {
		envList = buildEnv()
	}

	return envList
}

func globalenvs() map[string]interface{} {
	if (nil == globalEnvList) {
		globalEnvList = buildGlobalEnv()
	}

	return globalEnvList
}

func buildEnv () (map[string]interface{}) {
	env := make(map[string]interface{})
	var key string
	var value interface{}
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		key = pair[0]
		value = pair[1]
		if (!strings.HasPrefix(key, *envPrefix)) {
			continue
		}
		env[slugify(strings.TrimPrefix(strings.TrimPrefix(key, *envPrefix), `_`))] = formatRawValue(value)
	}

	return env
}

func buildGlobalEnv () (map[string]interface{}) {
	env := make(map[string]interface{})
	var key string
	var value interface{}
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		key = pair[0]
		value = pair[1]
		env[slugify(key)] = formatRawValue(value)
	}

	return env
}

func formatRawValue(v interface{}) (interface{}){
	return v
}

func slugify(v string) (string) {
	var r = regexp.MustCompile(`[^a-z0-9]+`)
	var r2 = regexp.MustCompile(`\s+`)
	return r2.ReplaceAllString(strings.TrimSpace(r.ReplaceAllString(strings.TrimSpace(strings.ToLower(v)), ` `)), `-`)
}

func snakize(v string) (string) {
	var r = regexp.MustCompile(`[^a-z0-9]+`)
	var r2 = regexp.MustCompile(`\s+`)
	return r2.ReplaceAllString(strings.TrimSpace(r.ReplaceAllString(strings.TrimSpace(strings.ToLower(v)), ` `)), `_`)
}

func underscore(v string) (string) {
	var camel = regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")

	var a []string
	for _, sub := range camel.FindAllStringSubmatch(v, -1) {
		if sub[1] != "" {
			a = append(a, sub[1])
		}
		if sub[2] != "" {
			a = append(a, sub[2])
		}
	}

	var vv string = strings.ToLower(strings.Join(a, "_"))
	var r = regexp.MustCompile(`[^a-z0-9]+`)
	var r2 = regexp.MustCompile(`\s+`)
	return r2.ReplaceAllString(strings.TrimSpace(r.ReplaceAllString(strings.TrimSpace(strings.ToLower(vv)), ` `)), `_`)
}

func envname (v string) (string) {
	words := strings.Fields(v)
	f := func (w string) string {
		return strings.Title(strings.ToLower(w))
	}
	for i, v := range words {
		words[i] = f(v)
	}

	return strings.ToUpper(underscore(strings.Join(words, ``)))
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

func envvalue(key string) interface{} {
	var localEnvs map[string]interface{} = envs()
	key = slugify(key)
	if val, ok := localEnvs[key]; ok {
		return val
	}

	return nil
}
func globalenvvalue(key string) interface{} {
	var localGlobalEnvs map[string]interface{} = globalenvs()
	key = slugify(key)
	if val, ok := localGlobalEnvs[key]; ok {
		return val
	}

	return nil
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
