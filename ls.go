package main

import (
	"container/list"
	"fmt"
	"log"
	"os"
	"os/user"
	"sort"
	"strings"
	"syscall"

	"github.com/logrusorgru/aurora"
	"github.com/ryanuber/columnize"
)

//-R
var recursive = false

//-l
var showList = false

//-i
var showInode = false

//-a
var showAll = false

//-d
var dirOnly = false

//-r
var reverse = false

//-U
var unordered = false

//Columnize configuration
var colConfig = columnize.DefaultConfig()

func totalBlocks(dir string) int {
	total := 0
	file, err := os.Open(dir)
	if err != nil {
		return 0
	}
	fiList, err1 := file.Readdir(0)
	if err1 != nil {
		return 0
	}
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	for _, fi := range fiList {
		if !showAll {
			if fi.Name()[0] == '.' {
				continue
			}
		}
		stata := getStat(dir + fi.Name())
		total += int(stata.Blocks / 2)
	}
	if showAll {
		total += 8
	}
	return total
}

func getStat(file string) syscall.Stat_t {
	var stat syscall.Stat_t
	if err := syscall.Stat(file, &stat); err != nil {
		log.Fatal(err)
	}
	return stat
}

func printList(dir *os.File, base string) (out string) {
	path := dir.Name() + "/" + base
	/* cfile, err1 := os.Open(path)
	if err1 != nil {
		log.Fatal(err1)
	} */
	var stat syscall.Stat_t
	if err2 := syscall.Stat(path, &stat); err2 != nil {
		log.Fatal(err2)
	}
	cfinfo, err3 := os.Stat(path)
	if err3 != nil {
		log.Fatal(err3)
	}
	if showInode {
		out += fmt.Sprintf("%v | ", stat.Ino)
	}
	out += fmt.Sprintf("%v | %v | ", cfinfo.Mode(), stat.Nlink)
	//print username and group name
	fu, err := user.LookupId(fmt.Sprint(stat.Uid))
	if err != nil {
		log.Fatal(err)
	}
	fg, err := user.LookupGroupId(fmt.Sprint(stat.Gid))
	if err != nil {
		log.Fatal(err)
	}
	out += fmt.Sprintf("%v | %v | %v | %v | ", fu.Username, fg.Name, cfinfo.Size(), cfinfo.ModTime().Format("Jan  2 15:04"))
	//print file name
	if cfinfo.IsDir() {
		out += fmt.Sprint(aurora.Bold(aurora.Blue(base)))
	} else {
		out += base
	}
	return out
}

func walk(file *os.File, prefix string) {
	/*fi, err1 := file.Stat()
	if err1 != nil {
		log.Fatal(err1)
	}*/
	if dirOnly {
		if showList {
			printList(file, ".")
		} else {
			fmt.Println(aurora.Bold(aurora.Blue(prefix)))
		}
		return
	}
	if recursive {
		fmt.Println(prefix + ":")
	}
	if showList {
		total := totalBlocks(file.Name())
		fmt.Printf("total %v\n", total)
	}
	dirInfo, err := file.Readdir(0)
	if err != nil {
		log.Fatal(err)
	}
	dirSize := len(dirInfo)
	children := make([]string, dirSize)
	cdirs := list.New()
	isDir := map[string]bool{}
	for i, cinfo := range dirInfo {
		if !showAll {
			if cinfo.Name()[0] == '.' {
				continue
			}
		}
		children[i] = cinfo.Name()
		if cinfo.IsDir() {
			cdirs.PushBack(cinfo.Name())
			isDir[cinfo.Name()] = true
		}
	}
	if showAll {
		children = append(children, ".")
		children = append(children, "..")
	}

	if !unordered {
		if reverse {
			sort.Sort(sort.Reverse(sort.StringSlice(children)))
		} else {
			sort.Sort(sort.StringSlice(children))
		}
	}

	if showList {
		lines := make([]string, 0)
		for _, base := range children {
			if len(base) == 0 {
				continue
			}
			lines = append(lines, printList(file, base))
		}
		result := columnize.Format(lines, colConfig)
		fmt.Print(result)
	} else {
		for _, base := range children {
			if len(base) == 0 {
				continue
			}
			if isDir[base] || base == "." || base == ".." {
				fmt.Print(aurora.Bold(aurora.Blue(base)))
			} else {
				fmt.Print(base)
			}
			fmt.Print("  ")
		}
	}
	fmt.Println("")

	if recursive {
		fmt.Println("")
		for i := cdirs.Front(); i != nil; i = i.Next() {
			fn := prefix + "/" + i.Value.(string)
			f1, err2 := os.Open(fn)
			if err2 != nil {
				log.Fatal(err2)
			}
			walk(f1, fn)
		}
	}

}

func handle(dir, prefix string) {
	dirFile, err1 := os.Open(dir)
	if err1 != nil {
		log.Fatal(err1)
	}
	walk(dirFile, prefix)
}

func parseArg(ch rune) {
	switch ch {
	case 'a':
		showAll = true
	case 'd':
		dirOnly = true
	case 'i':
		showInode = true
	case 'l':
		showList = true
	case 'R':
		recursive = true
	case 'r':
		reverse = true
	case 'U':
		unordered = true
	}
}

func main() {
	colConfig.Glue = " "

	args := os.Args[1:]
	dir := ""
	prefix := "."
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	if len(args) > 0 {
		lastArg := args[len(args)-1]
		if lastArg[0] == '/' {
			dir = lastArg
			prefix = lastArg
			if prefix[len(prefix)-1] == '/' {
				prefix = prefix[:len(prefix)-1]
			}
		} else if lastArg[0] == '.' {
			dir = wd + "/" + lastArg
			prefix = lastArg
		}
	}
	for _, arg := range args {
		if arg[0] == '-' {
			for _, ch := range arg[1:] {
				parseArg(ch)
			}
		}
	}
	handle(dir, prefix)
}
