package typescript

import (
	"fmt"
	"log"
	"mime"
	"os"
	"os/exec"

	"github.com/nonanick/impatience/transform"
)

// TsConverterScriptName name of the file that will be created
var TsConverterScriptName = "impatience-ts-transpiler"

// Register Typescript transformer
func Register() {
	mime.AddExtensionType(".ts", "text/javascript")
	transform.AddFileTransformer(".ts", TranspileTs)
}

var transpilerExists = false

//TranspileTs transpile a ts file generating an in memory js file
var TranspileTs transform.FileTransformer = func(path string, content []byte) []byte {

	if !transpilerExists {
		generateTsConverterScript()
		transpilerExists = true
	}

	// This works, lets try output pipe
	cmd := exec.Command("node", TsConverterScriptName, path)

	out, outErr := cmd.Output()
	if outErr != nil {
		fmt.Println("Failed to obtain output from transpiler!")
		return content
	}

	fmt.Println("TS transpilation finished for file: ", path)
	return out
}

func generateTsConverterScript() {

	transpiler, err := os.Open(TsConverterScriptName)
	if err != nil {
		fmt.Println("Could not find transpiler")

		emptyTrans, createErr := os.Create(TsConverterScriptName)
		if createErr != nil {
			log.Fatal("Could not create Typescript transpiler!")
		}

		emptyTrans.Write([]byte(transpilerScriptContent))

		defer emptyTrans.Close()
	}

	defer transpiler.Close()

}

var transpilerScriptContent = `
const ts =  require('typescript');
const fs = require('fs')

const args = process.argv

if (args.length > 2) {
	
	const compilepath = args[2];
	const content = fs.readFileSync(compilepath).toString()
	const transpiled = ts.transpile(content, {
		module : ts.ModuleKind.ES2015,
		target : ts.ScriptTarget.ES2015,
		esModuleInterop : true,
		noEmit : true,
		experimentalDecorators : true,
		moduleResolution : ts.ModuleResolutionKind.NodeJs,
		skipLibCheck : true,
		inlineSourceMap : true,
		transpileOnly : true
	})

	process.stdout.write(transpiled);
} 
`
