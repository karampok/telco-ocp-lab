package main

import (
	"embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/karampok/telco-ocp-lab/pkg"
	"github.com/saschagrunert/demo"
)

//go:embed config/*
var configFS embed.FS

//go:embed infra/*
var infraFS embed.FS

//go:embed topo.clab.yaml
var cclab []byte

//go:embed vbmh-kcli-plan.yaml
var kplan []byte

func main() {
	d := demo.New()

	d.Name = "telco-ocp-lab"
	d.Description = "Setup virtual infra for multi-interface cluster"

	d.Add(pkg.SetupInfra(), "setup", "setup virtual infra")
	d.Add(pkg.Clean(), "clean", "clean system")
	d.Add(pkg.RunIPForwardingDemo(), "ipforwarding", "reproduce ipforwarding demo")
	d.Add(pkg.RunBGPGracefulRestart(), "BGP-GR", "demo BGP w,w/o GR (Graceful restart)")
	d.Add(pkg.RunBGPGracefulRestartWithBFD(), "BGP-GR-BFD", "demo BGP w,w/o GR (Graceful restart), BFD")

	if err := extractConfig(); err != nil {
		os.Exit(1)
	}

	d.Run()
}

func extractConfig() error {
	clab := "topo.clab.yaml"
	_, err := os.Stat(clab)
	if os.IsNotExist(err) {
		if err := os.WriteFile(clab, cclab, 0o644); err != nil {
			return err
		}
	}

	plan := "vbmh-kcli-plan.yaml"
	_, err = os.Stat(plan)
	if os.IsNotExist(err) {
		if err := os.WriteFile(plan, kplan, 0o644); err != nil {
			return err
		}
	}

	extractDir := func(efs *embed.FS) error {
		files, err := getAllFilenames(efs)
		if err != nil {
			return err
		}
		for _, f := range files {
			src, err := efs.Open(f)
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(f), 0o755); err != nil {
				return err
			}

			dst, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0o644)
			if err != nil {
				return err
			}
			if _, err = io.Copy(dst, src); err != nil {
				return err
			}
		}
		return nil
	}
	for _, fs := range []*embed.FS{&configFS, &infraFS} {
		if err := extractDir(fs); err != nil {
			return err
		}
	}
	return nil
}

func getAllFilenames(efs *embed.FS) (files []string, err error) {
	if err := fs.WalkDir(efs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		//		if _, err := os.Stat(path); os.IsNotExist(err) {
		files = append(files, path)
		//		}

		return nil
	}); err != nil {
		return nil, err
	}

	return files, nil
}
