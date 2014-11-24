package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/coreos-inc/rkt/stage0"
)

var (
	flagStage1Init   string
	flagStage1Rootfs string
	flagVolumes      volumeMap
	cmdRun           = &Command{
		Name:    "run",
		Summary: "Run image(s) in an application container in rocket",
		Usage:   "[--volume LABEL:SOURCE] IMAGE...",
		Run:     runRun,
	}
)

func init() {
	cmdRun.Flags.StringVar(&flagStage1Init, "stage1-init", "./bin/init", "path to stage1 binary")
	cmdRun.Flags.StringVar(&flagStage1Rootfs, "stage1-rootfs", "./stage1-rootfs.tar.gz", "path to stage1 rootfs tarball")
	cmdRun.Flags.Var(&flagVolumes, "volume", "volumes to mount into the shared container environment")
	flagVolumes = volumeMap{}
}

func runRun(args []string) (exit int) {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "run: Must provide at least one image\n")
		return 1
	}
	gdir := globalFlags.Dir
	if gdir == "" {
		log.Printf("dir unset - using temporary directory")
		var err error
		gdir, err = ioutil.TempDir("", "rkt")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating temporary directory: %v", err)
			return 1
		}
	}
	// TODO(jonboulle): use rkt/path
	cdir := filepath.Join(gdir, "containers")
	cfg := stage0.Config{
		ContainersDir: cdir,
		Debug:         globalFlags.Debug,
		Stage1Init:    flagStage1Init,
		Stage1Rootfs:  flagStage1Rootfs,
		Images:        args,
		Volumes:       flagVolumes,
	}
	cdir, err := stage0.Setup(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "run: error setting up stage0: %v\n", err)
		return 1
	}
	stage0.Run(cdir, cfg.Debug) // execs, never returns
	return 1
}

// volumeMap implements the flag.Value interface to contain a set of mappings
// from mount label --> mount path
type volumeMap map[string]string

func (vm *volumeMap) Set(s string) error {
	elems := strings.Split(s, ":")
	if len(elems) != 2 {
		return errors.New("volume must be of form key:path")
	}
	key := elems[0]
	if _, ok := (*vm)[key]; ok {
		return fmt.Errorf("got multiple flags for volume %q", key)
	}
	(*vm)[key] = elems[1]
	return nil
}

func (vm *volumeMap) String() string {
	var ss []string
	for k, v := range *vm {
		ss = append(ss, fmt.Sprintf("%s:%s", k, v))
	}
	return strings.Join(ss, ",")
}
