package main

import (
	"flag"
	"fmt"
)

var (
	sourceTplFile = flag.String("s", "file.tpl", "absolute path to the source template file")
	targetFile    = flag.String("t", "file.conf", "absolute path to the target file generated")
	envPrefix     = flag.String("p", "APPCONF", "env var prefix")
)

func main() {
	flag.Parse()
	fmt.Println("Source Tpl is  :", *sourceTplFile)
	fmt.Println("targetFile is : ", *targetFile)
	fmt.Println("Env Prefix : ", *envPrefix)
}
