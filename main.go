package main

import (
	"multifs/config"
	"multifs/mergednode"
	"multifs/pathiterator"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/log"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

func main() {
	log.SetReportTimestamp(false)
	log.SetPrefix("MULTIFS")
	log.SetReportCaller(true)

	config.InitializeConfig()

	// debug := flag.Bool("debug", false, "print debug data")
	// flag.Parse()
	//
	// if *debug {
	// 	log.SetLevel(log.DebugLevel)
	// }

	userConfig := config.GetConfig()

	var rootPaths []string
	mntDir := userConfig.MountDir
	debug := userConfig.Debug

	if debug {
		log.SetLevel(log.DebugLevel)
	}

	for _, p := range userConfig.RootPaths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			log.Fatal("Error in path", p, err)
		}
		rootPaths = append(rootPaths, absPath)
	}
	log.Printf("%v %v %v", rootPaths, mntDir, debug)

	log.Info("Mounting", "paths", rootPaths)

	os.Mkdir(mntDir, 0755)

	pathiterator.NewTree(rootPaths, mntDir)
	root := mergednode.NewMergedNode()

	sec := time.Second
	opts := &fs.Options{
		// The timeout options are to be compatible with libfuse defaults,
		// making benchmarking easier.
		// AttrTimeout:  &sec,
		EntryTimeout: &sec,

		MountOptions: fuse.MountOptions{
			// Set to true to see how the file system works.
			Debug: false,
		},
	}

	server, err := fs.Mount(mntDir, root, opts)
	if err != nil {
		log.Error(err)
	}

	log.Infof("Mounted on %s", mntDir)
	log.Infof("Unmount by calling 'fusermount -u %s'", mntDir)

	server.Wait()

	os.Remove(mntDir)
}
