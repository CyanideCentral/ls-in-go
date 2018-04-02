package main

import(
	"fmt"
	"os"
	"log"
)

//-R
var recursive bool = false

//-l
var list bool = false

//-i
var show_inode bool = false

//-a
var show_all bool = false

//-d
var dir_only = false

func handle(dir string){
	dirFile, err1 := os.Open(dir)
	if err1 != nil {
		log.Fatal(err1)
	}

	fmt.Println(dirFile.Stat())
}

func main(){
	args := os.Args[1:]
	dir := ""
	if len(args) > 0{
		lastArg := args[len(args)-1]
		if lastArg[0] == '/' {
			dir = lastArg
		}
	}
	if dir == "" {
		dir, _ = os.Getwd()
	}
	handle(dir)
}