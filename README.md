# dailyprog

This program can be used to quickly get started with a "program of the day" in the Go language.

It does the following:
* Create a directory named ```<homedirectory>prog_of_the_day/yyyymmdd```
* Create the standard files to get going (for go: main.go, go.mod, go.sum, vscode debug launch fiule)
* Init Git
* Open new folder in Visual Studio Code

WARNING: any files already present in ```<homedirectory>prog_of_the_day/yyyymmdd``` may be overwritten!

To install **dailyprog** from source:
* Install [Go](https://go.dev/dl/)
* Execute (in a shell windows): ```go install github.com/524D/dailyprog@latest```

The ```dailyprog``` executable will now be in ```<homedirectory>/go/bin/dailyprog```

