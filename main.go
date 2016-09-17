package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var typeMap = map[byte]string{
	'0': "f", // regular file
	0:   "f", // regular file
	'1': "f", // hard link represented as regular file (genext2fs will see the link in the real filesystem)
	'2': "f", // symlink represented as regular file (genext2fs will see the link in the real filesystem)
	'3': "c", // character device
	'4': "b", // block device
	'5': "d", // directory
	'6': "p", // FIFO
}

// Types where we need to populate the major/minor numbers
var deviceTypes = map[byte]bool{
	'3': true, // character device
	'4': true, // block device
}

func main() {
	flag.Usage = help
	flag.Parse()

	args := flag.Args()

	err := run(args)

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(2)
	}

	os.Exit(0)
}

func run(args []string) error {
	var f *os.File
	var err error
	if len(args) == 0 {
		f = os.Stdin
	} else if len(args) == 1 {
		f, err = os.Open(args[0])
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("extraneous arguments; use -help for usage")
	}

	reader := tar.NewReader(f)

	for {
		hdr, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		absPath := filepath.Join("/", filepath.Clean(hdr.Name))
		if absPath == "/" {
			// Ignore any entry that describes the root
			continue
		}
		typeCode, typeOk := typeMap[hdr.Typeflag]
		if !typeOk {
			fmt.Fprintf(os.Stderr, "ignoring %s: unsupported typeflag %x\n", absPath, hdr.Typeflag)
			continue
		}

		// Start of every line is common, regardless of type
		fmt.Printf(
			"%s %s %03o %d %d ",
			absPath,
			typeCode,
			hdr.Mode,
			hdr.Uid,
			hdr.Gid,
		)
		if deviceTypes[hdr.Typeflag] {
			// major/minor output for devices
			fmt.Printf("%d %d 0 0 -\n", hdr.Devmajor, hdr.Devminor)
		} else {
			// placeholders for non-devices
			fmt.Printf("- - - - -\n")
		}
	}

	return nil
}

func help() {
	fmt.Fprintf(os.Stderr, "\nUsage: %s [tar-file]\n\n", os.Args[0])
	fmt.Fprintf(
		os.Stderr,
		"Will read from stdin if no tar-file argument is provided\n\n",
	)
	flag.PrintDefaults()
}
