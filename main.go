package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/dustin/go-humanize"
	"github.com/urfave/cli/v2"
)

type metadata struct {
	dMeta dirMap
	fMeta fileMap
	total int64
}

type (
	dirMap  map[string]dirMeta
	fileMap map[string]int64
)

type dirMeta struct {
	size         int64
	bytesDeleted int64
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
		Before: func(ctx *cli.Context) error {
			args := ctx.Args()
			if args.Len() != 1 {
				return errors.New("error invalid args: exactly one argument must be provided")
			}
			return nil
		},
		Action: func(ctx *cli.Context) error {
			exts := ctx.StringSlice("ext")
			rootDir := ctx.Args().Get(0)

			meta, err := collectDirMetadata(rootDir, exts)
			if err != nil {
				return err
			}

			if meta.total == 0 {
				fmt.Println("There is nothing to delete. Exiting...")
				return nil
			}

			if err := meta.reportFileMetadata(); err != nil {
				return err
			}

			confirm := askForConfirmation("do you want to go ahead with deleting these files?")

			if !confirm {
				fmt.Println("exiting...")
				return nil
			}

			meta, err = deleteFilesByExtension(rootDir, meta)
			if err != nil {
				return err
			}

			if err := meta.reportDirMetadata(); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func (d dirMap) report() error {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "DIRECTORY\tOLDSIZE\tNEWSIZE\tBYTES SAVED\n")
	fmt.Fprint(w, "---------\t-------\t-------\t-----------\n")
	for k, v := range d {
		if v.bytesDeleted != 0 {
			size := humanize.Bytes(uint64(v.size))
			newsz := humanize.Bytes(uint64(v.size - v.bytesDeleted))
			bytesSaved := humanize.Bytes(uint64(v.bytesDeleted))

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				k,
				size,
				newsz,
				bytesSaved,
			)
		}
	}
	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}

	fmt.Print("\n")
	return nil
}

func (m metadata) reportFileMetadata() error {
	if m.total == 0 {
		return nil
	}

	return m.fMeta.report(m.total)
}

func (m metadata) reportDirMetadata() error {
	return m.dMeta.report()
}

func (f fileMap) report(total int64) error {
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "FILE\tSIZE\n")
	fmt.Fprint(w, "----\t----\n")
	for k, v := range f {
		fmt.Fprintf(w, "%s\t%s\n", k, humanize.Bytes(uint64(v)))
	}

	fmt.Fprint(w, "----\t----\n")
	fmt.Fprintf(w, "TOTAL\t%s\n\n", humanize.Bytes(uint64(total)))

	if err := w.Flush(); err != nil {
		log.Fatal(err)
	}

	return nil
}

func deleteFilesByExtension(dir string, meta metadata) (metadata, error) {
	for path, size := range meta.fMeta {
		dir := filepath.Dir(path)
		sz, ok := meta.dMeta[dir]
		if ok {
			err := os.Remove(path)
			if err != nil {
				return metadata{}, err
			}
			sz.bytesDeleted += size
			meta.dMeta[dir] = sz
		}
	}

	return meta, nil
}

func collectDirMetadata(rootdir string, exts []string) (metadata, error) {
	dmap := make(dirMap)
	fmap := make(fileMap)
	var total int64

	err := filepath.Walk(rootdir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			var d dirMeta
			dmap[path] = d
		}

		if !info.IsDir() {
			if matchExt(info.Name(), exts) {
				size := info.Size()
				fmap[path] = size
				total += size
			}

			dir := filepath.Dir(path)
			sz, ok := dmap[dir]
			if !ok {
				sz.size += info.Size()
				dmap[dir] = sz
			}
		}

		return nil
	})
	if err != nil {
		return metadata{}, err
	}

	return metadata{
		dMeta: dmap,
		fMeta: fmap,
		total: total,
	}, nil
}

func matchExt(file string, ext []string) bool {
	for _, e := range ext {
		if strings.TrimLeft(filepath.Ext(file), ".") == e {
			return true
		}
	}
	return false
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		fmt.Print("\n")

		switch response {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}
	}
}
