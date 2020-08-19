package main

import "github.com/kr/pretty"

// Help display the help options
func Help(args []string) {
	pretty.Println(
		"[Impatience - Help]:\n",
		"	- Avaliable sub commands:\n",
		"		º launch\n",
		"		º init\n",
		"		º help\n",
		"------------------------------------------\n\n",
		// Launch
		"# command \"launch\": \n",
		"Launches a new web server.\n",
		"	--address, -a  server address\n",
		"	--cache, -s    cache strategy, as of now only \"cookie\" is valid\n",
		"	--config, -c   path for a JSON configuration\n",
		"	--node, -n     path to node_modules root\n",
		"	--node-ext     file extensions that shall be analyzed looking for node libraries\n",
		"	--port, -p     TCP port the server shall be launched in\n",
		"	--root, -r     public root that shall be served by Impatience\n",
		"	--ts           enable ts support, you may specify the path to tsconfig\n",
	)
}
