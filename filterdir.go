package webapps

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"

	"golang.org/x/net/context"

	"github.com/jroimartin/gocui"
	"github.com/shurcooL/vfsgen"
)

// FilterDir is a http.FileSystem middleware that implements filtering of
// files, which restrict visible files to those in the IncludeList.
type FilterDir struct {
	loadOnce  sync.Once
	startOnce sync.Once
	dir       http.Dir
	options   Options
	requests  chan string
	include   map[string]struct{}

	// ProdMode enables the filter so only files found in IncludeList
	// will be returned.
	ProdMode bool

	// IncludeList is a slice of files that are allowed to be returned when
	// ProdMode is set to true.
	IncludeList []string
}

// NewFilterDir returns a newly instanciated FilterDir with dir as the root directory used to server files.
func NewFilterDir(dir string, opt Options) *FilterDir {
	opt.fillMissing()
	return &FilterDir{dir: http.Dir(dir), options: opt, requests: make(chan string, 100)}
}

// Options used by vfsgen when generating the statically implemented virtual filesystem.
func (f *FilterDir) Options() vfsgen.Options {
	return vfsgen.Options{
		Filename:        f.options.Filename,
		PackageName:     f.options.PackageName,
		BuildTags:       f.options.VfsgenBuildTags,
		VariableName:    f.options.VariableName,
		VariableComment: f.options.VariableComment,
	}
}

// Open attempts to open name, which is a resource under the root dir provided to FilterDir
func (f *FilterDir) Open(name string) (http.File, error) {
	file, err := f.dir.Open(name)
	if err != nil {
		return nil, err
	}
	if f.ProdMode == false {
		f.startOnce.Do(func() {
			go f.startGUI()
		})
		f.requests <- name
		return file, nil
	}

	// We are in ProdMode, so results will be filtered
	f.loadOnce.Do(f.loadIncludeList)

	if _, ok := f.include[name]; ok {
		return &File{File: file, name: name, include: f.include}, nil
	}

	return nil, os.ErrNotExist
}

func (f *FilterDir) loadIncludeList() {
	f.include = make(map[string]struct{})
	f.include["/"] = struct{}{}
	for _, file := range f.IncludeList {
		f.include[file] = struct{}{}
		dirs := strings.Split(file, "/")
		for i := 2; i < len(dirs); i++ {
			f.include[strings.Join(dirs[:i], "/")] = struct{}{}
		}
	}
}

func (f *FilterDir) startGUI() {
	reqs := processRequests(f.IncludeList, f.requests)
	screen := gocui.NewGui()
	if err := screen.Init(); err != nil {
		log.Fatal(err)
	}
	defer screen.Close()
	screen.SetLayout(layout)
	screen.Cursor = true

	ctx, done := context.WithCancel(context.Background())
	defer done()

	go pushUpdates(ctx, screen, reqs)

	if err := bindKeys(screen, reqs, f); err != nil {
		log.Fatal(err)
	}
	if err := screen.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatal(err)
	}
}

func (f *FilterDir) generateAssets(list []string) error {
	f.IncludeList = list
	f.ProdMode = true

	err := vfsgen.Generate(f, f.Options())
	if err != nil {
		return err
	}
	return nil
}

func (f *FilterDir) saveList(list []string) error {

	// Create output file.
	lf, err := os.Create(f.options.ListFileName)
	if err != nil {
		return err
	}
	defer lf.Close()

	err = t.ExecuteTemplate(lf, "Header", f.options)
	if err != nil {
		return err
	}

	for _, l := range list {
		err = t.ExecuteTemplate(lf, "Files", l)
		if err != nil {
			return err
		}
	}

	err = t.ExecuteTemplate(lf, "Footer", f.options)
	if err != nil {
		return err
	}

	return nil
}

func processRequests(savedIncludeList []string, requests chan string) *sortedList {
	var changed bool
	qchan := make(chan struct{})
	achan := make(chan []string)
	cchan := make(chan bool)
	clear := make(chan struct{})

	includeList := make([]string, 0, 100)
	includeMap := make(map[string]bool)

	go func() {
		for {
			select {
			case r := <-requests:
				if includeMap[r] == false {
					includeMap[r] = true
					includeList = append(includeList, r)
					changed = true
				}
			case <-qchan:
				// sort includeList
				if changed {
					sort.StringSlice(includeList).Sort()
					changed = false
				}
				sortedList := make([]string, len(includeList))
				copy(sortedList, includeList)
				achan <- sortedList
			case cchan <- changed:
			case <-clear:
				includeList = make([]string, 0, 100)
				includeMap = make(map[string]bool)
				changed = true
			}
		}
	}()

	return &sortedList{question: qchan, answer: achan, changed: cchan, clear: clear}
}

type sortedList struct {
	question chan struct{}
	answer   chan []string
	changed  chan bool
	clear    chan struct{}
}

func (l *sortedList) List() []string {
	l.question <- struct{}{}
	return <-l.answer
}

func (l *sortedList) Changed() bool {
	return <-l.changed
}

func (l *sortedList) Clear() {
	l.clear <- struct{}{}
}

// A File is returned by a FileSystem's Open method and can be
// served by the FileServer implementation.
//
// The methods should behave the same as those on an *os.File.
type File struct {
	http.File
	name    string
	include map[string]struct{}
}

// Readdir behaves the same way as os.File.Readdir, but additionally
// filters on IncludeList
func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	// Remove trailing '/' if it is present
	if f.name[len(f.name)-1:] == "/" {
		f.name = f.name[:len(f.name)-1]
	}
	info, err := f.File.Readdir(count)
	var newInfo []os.FileInfo
	for _, i := range info {
		if _, ok := f.include[f.name+"/"+i.Name()]; ok {
			newInfo = append(newInfo, i)
		}
	}
	return newInfo, err
}

// Options for vfsgen code generation.
type Options struct {
	// Filename of the generated Go code output (including extension).
	// If left empty, it defaults to "{{toLower .VariableName}}_vfsdata.go".
	Filename string

	// PackageName is the name of the package in the generated code.
	// If left empty, it defaults to "main".
	PackageName string

	// VfsgenBuildTags are the optional build tags in the generated code.
	// If left empty, it defaults to "!dev".
	// The build tags syntax is specified by the go tool.
	VfsgenBuildTags string

	// VariableName is the name of the http.FileSystem variable in the generated code.
	// If left empty, it defaults to "assets".
	VariableName string

	// VariableComment is the comment of the http.FileSystem variable in the generated code.
	// If left empty, it defaults to "{{.VariableName}} statically implements the virtual filesystem provided to vfsgen.".
	VariableComment string

	// ListFileName is the name of the go source file which holds the generated code for IncludeList.
	// If left empty, it defaults to "assets_list.go".
	ListFileName string

	// ListFileBuildTags are the optional build tags in the generated code for IncludeList.
	// If left empty, it defaults to "dev".
	// The build tags syntax is specified by the go tool.
	ListFileBuildTags string
}

// fillMissing sets default values for mandatory options that are left empty.
func (opt *Options) fillMissing() {
	if opt.PackageName == "" {
		opt.PackageName = "main"
	}
	if opt.VariableName == "" {
		opt.VariableName = "assets"
	}
	if opt.ListFileName == "" {
		opt.ListFileName = "assets_list.go"
	}
	if opt.VfsgenBuildTags == "" {
		opt.VfsgenBuildTags = "!dev"
	}
	if opt.ListFileBuildTags == "" {
		opt.ListFileBuildTags = "dev"
	}
}

var t = template.Must(template.New("").Parse(`{{define "Header"}}// Code generated by FilterDir

{{with .ListFileBuildTags}}// +build {{.}}

{{end}}package {{.PackageName}}

func init() {
	{{.VariableName}}.IncludeList = []string{
{{end}}

{{define "Files"}}	"{{.}}",
{{end}}

{{define "Footer"}}	}
}
{{end}}
`))
