package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/dustin/go-humanize"
	"github.com/urfave/cli/v2"
)

type Metadata struct {
	Dirs  map[string]*DirMeta
	Files map[string]int64
	Total int64
}

type DirMeta struct {
	Size    int64
	Deleted int64
}

type Reporter interface {
	Report() error
}

type FileReporter struct {
	Files map[string]int64
	Total int64
}

func (f *FileReporter) Report() error {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)

	fmt.Fprint(w, "FILE\tSIZE\n")
	fmt.Fprint(w, "----\t----\n")

	for path, size := range f.Files {
		fmt.Fprintf(w, "%s\t%s\n", path, humanize.Bytes(uint64(size)))
	}

	fmt.Fprint(w, "----\t----\n")
	fmt.Fprintf(w, "TOTAL\t%s\n", humanize.Bytes(uint64(f.Total)))

	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}

type DirReporter struct {
	Dirs map[string]*DirMeta
}

func (d *DirReporter) Report() error {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)

	fmt.Fprint(w, "DIRECTORY\tOLDSIZE\tNEWSIZE\tBYTES SAVED\n")
	fmt.Fprint(w, "---------\t-------\t-------\t-----------\n")

	for dir, meta := range d.Dirs {
		if meta.Deleted > 0 {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				dir,
				humanize.Bytes(uint64(meta.Size)),
				humanize.Bytes(uint64(meta.Size-meta.Deleted)),
				humanize.Bytes(uint64(meta.Deleted)),
			)
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}

	return nil
}

func WalkDirs(root string, exts []string) chan Metadata {
	outChan := make(chan Metadata)

	go func() {
		defer close(outChan)

		meta := Metadata{
			Dirs:  make(map[string]*DirMeta),
			Files: make(map[string]int64),
		}

		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			dir := filepath.Dir(path)

			if !info.IsDir() {
				ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(info.Name()), "."))
				if matchExt(ext, exts) {
					meta.Files[path] = info.Size()
					meta.Total += info.Size()
				}
			}

			dm, ok := meta.Dirs[dir]
			if !ok {
				dm = &DirMeta{}
			}
			dm.Size += info.Size()
			meta.Dirs[dir] = dm

			return nil
		})

		if err != nil {
			log.Fatal(err)
		}

		outChan <- meta
	}()

	return outChan
}

func deleteFiles(meta Metadata) chan Metadata {
	outChan := make(chan Metadata)

	go func() {
		defer close(outChan)

		var wg sync.WaitGroup
		var filesToDelete []string

		for path := range meta.Files {
			filesToDelete = append(filesToDelete, path)
		}

		if len(filesToDelete) == 0 {
			fmt.Println("Nothing to delete")
			outChan <- meta
			return
		}

		var mu sync.Mutex
		var totalDeleted int64

		for _, path := range filesToDelete {
			wg.Add(1)

			go func(path string) {
				defer wg.Done()

				err := os.Remove(path)
				if err != nil {
					log.Printf("Error deleting %s: %v", path, err)
					return
				}

				dir := filepath.Dir(path)

				mu.Lock()
				dm, ok := meta.Dirs[dir]
				if ok {
					dm.Deleted += meta.Files[path]
					meta.Dirs[dir] = dm
				}

				delete(meta.Files, path)

				for dir != "." && dir != "/" {
					dm, ok := meta.Dirs[dir]
					if ok {
						dm.Deleted += meta.Files[path]
						meta.Dirs[dir] = dm
					}
					dir = filepath.Dir(dir)
				}
				mu.Unlock()

				totalDeleted += meta.Files[path]
			}(path)
		}

		wg.Wait()

		meta.Total -= totalDeleted

		outChan <- meta
	}()

	return outChan
}

func main() {
	app := &cli.App{
		Usage:           "Delete files within a directory structure by file extensions",
		UsageText:       "delly [global options] command [arguments...]",
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:     "ext",
				Aliases:  []string{"e"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				return fmt.Errorf("Usage: delly [global options] <directory>")
			}

			rootDir := c.Args().Get(0)
			exts := c.StringSlice("ext")

			metaChan := WalkDirs(rootDir, exts)
			meta := <-metaChan

			if meta.Total == 0 {
				fmt.Println("Nothing to delete")
				return nil
			}

			reporter := &FileReporter{meta.Files, meta.Total}
			if err := reporter.Report(); err != nil {
				log.Fatal(err)
			}

			confirm := askConfirm("Confirm delete? [y/n]: ")
			if !confirm {
				fmt.Println("Exiting...")
				return nil
			}

			delChan := deleteFiles(meta)
			meta = <-delChan

			dirReporter := &DirReporter{meta.Dirs}
			if err := dirReporter.Report(); err != nil {
				log.Fatal(err)
			}

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func askConfirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s", prompt)

		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		switch strings.ToLower(text) {
		case "y", "yes":
			return true

		case "n", "no":
			return false
		}
	}
}

func matchExt(ext string, exts []string) bool {
	ext = strings.ToLower(ext)
	for _, validExt := range exts {
		if ext == validExt {
			return true
		}
	}
	return false
}
