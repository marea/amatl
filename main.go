package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	s "strings"
	"time"
)

const SITENAME string = "marea"
const RESDIR string = "dist"
const INCDIR string = "inc"
const ASTDIR string = "assets"
const HEADER string = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<link rel="stylesheet" href="/style.css">
<title>%s - %s</title>
</head>
<body>
	<header><h1>` + SITENAME + `</h1></header>
	<main>
`
const FOOTER string = `
		%s
		<div>Last modified date: %s</div>
	</main>
<footer>
This website and it's contents are released under the
<a href="https://unlicense.org/">unlicense</a>, unless otherwise stated.
Do whatever you want ;)
</footer>
</body>`

type File struct {
	Name      string
	FileName  string
	Path      string
	Content   string
	LinksHere []File
	URI       string
	Info      fs.FileInfo
}

// Get file name without extension
func fnwe(f string) string { return s.TrimSuffix(f, filepath.Ext(f)) }

// Get file dir path
func fdp(f string) string { return filepath.Dir(f) }

// Get file content
func fgc(f string) string {
	c, err := os.ReadFile(f)
	handleError(err)
	return string(c)
}

// Get links in content
func cgl(c string) [][]byte {
	return regexp.MustCompile(`\<a.*\/a>`).FindAll([]byte(c), -1)
}

// Get URL in link
func lgu(l string) string {
	return s.Trim(s.Trim(string(regexp.MustCompile(
		`(?s)href=(["'])(.*?)(["'])`).Find([]byte(l))), "href="), `"`)
}

// Checks if link is external
func lie(l string) bool {
	return s.Index(l, "http://") == 0 || s.Index(l, "https://") == 0
}

// Get File path without INCDIR
func fwi(p string) string {
	return s.Trim(p, INCDIR+"/")
}

// Find File URI
func ffu(u File) string {
	return "/" + s.Trim(fwi(fdp(u.Path+"/"))+"/"+u.FileName, "/")
}

// Function that deals with files under the INCDIR
func wfn(files *[]File, backlinks *map[string][]string) func(path string,
	d fs.DirEntry, err error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			filename := d.Name()
			content := fgc(path)
			name := fnwe(filename)
			if s.Index(content, "# title:") == 0 {
				name = s.Split(s.Split(content, "\n")[0], ":")[1]
				content = s.Join(s.Split(content, "\n")[1:], "\n")
			}
			file := File{
				Name:     name,
				FileName: filename,
				Path:     fdp(path),
				Content:  content,
			}
			file.URI = ffu(file)
			file.Info, err = d.Info()
			handleError(err)

			for _, link := range cgl(fgc(path)) {
				link := string(link)
				index := lgu(link)
				(*backlinks)[index] = append((*backlinks)[index], fwi(path))
			}
			(*files) = append((*files), file)
			return nil
		}
		return err
	}
}

// Adds backlink to File
func (f *File) AddBackLink(linkedIn File) {
	f.LinksHere = append(f.LinksHere, linkedIn)
}

// Adds backlinks to all Files in array
func AddBacklinks(files *[]File, backlinks map[string][]string) {
	for index, current := range *files {
		path := fwi(current.Path) + "/" + current.FileName
		for _, link := range (backlinks)[path] {
			for _, f := range *files {
				if fwi(f.Path+"/"+f.FileName) == link {
					(*files)[index].AddBackLink(File{
						Name:     f.Name,
						FileName: f.FileName,
						Path:     fwi(f.Path),
						URI:      ffu(f),
					})
				}
			}
		}
	}
}

// Load Files in INCDIR
func GetFiles() []File {
	backlinks := make(map[string][]string, 0)
	var files []File

	err := filepath.WalkDir(INCDIR, wfn(&files, &backlinks))
	handleError(err)

	AddBacklinks(&files, backlinks)
	return files
}

// Format time with given layout
func tfmt(t time.Time, l string) string { return t.Format(l) }

// Generate backlinks list
func btl(b []File) string {
	var res s.Builder
	if len(b) > 0 {
		fmt.Fprintf(&res, "<ul class=\"backlinks\">")
		for _, file := range b {
			fmt.Fprintf(&res, "<li><a href=\"%s\">%s</a></li>", file.URI, file.Name)
		}
		fmt.Fprintf(&res, "</ul>")
	}
	return res.String()
}

// Saves file in RESTDIR
func SaveFile(f File) {
	path := filepath.Dir(RESDIR + f.URI)
	err := os.MkdirAll(path, 0777)
	handleError(err)
	err = os.WriteFile(path+"/"+f.FileName, []byte(f.Content), 0777)
	handleError(err)
}

// Compiles the given file
func (f File) Compile() {
	var res s.Builder
	fmt.Fprintf(&res, HEADER, SITENAME, f.Name)
	fmt.Fprintf(&res, f.Content)
	fmt.Fprintf(&res, FOOTER, btl(f.LinksHere), tfmt(f.Info.ModTime(),
		"2006/01/02"))
	f.Content = res.String()
	SaveFile(f)
}

// Copies files from ASTDIR to RESDIR
func CopyAssets() {
	err := filepath.WalkDir(ASTDIR, func(path string, d fs.DirEntry,
		err error) error {

		resPath := RESDIR + s.TrimLeft(path, ASTDIR)
		if !d.IsDir() {
			err := os.WriteFile(resPath, []byte(fgc(path)), 0777)
			handleError(err)
			return nil
		} else {
			err := os.MkdirAll(resPath, 0777)
			handleError(err)
			return nil
		}
	})
	handleError(err)
}

func checkDirectories() {
	for _, path := range []string{ASTDIR, RESDIR, INCDIR} {
		_, err := os.Stat(path)
		if err != nil {
			os.Mkdir(path, 0777)
		}
	}
}

func main() {
	checkDirectories()
	for _, file := range GetFiles() {
		file.Compile()
	}
	CopyAssets()
}

func handleError(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
}
