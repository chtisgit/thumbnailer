package main

import (
	"flag"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/chtisgit/thumbnailer/thmb"
)

func serve(configFile string, chmod bool) {
	var cfg thmb.Config
	_, err := toml.DecodeFile(configFile, &cfg)
	if err != nil {
		log.Printf("Cannot parse config file %s\n", configFile)
		log.Printf("error: %s\n", err)
		return
	}

	srv, err := thmb.NewServer(&cfg)
	if err != nil {
		log.Print(err)
		return
	}

	if chmod {
		os.Chmod(cfg.Addr, 0777)
	}

	log.Println("Ready.")
	err = srv.Serve(nil)
	log.Print(err)
}

func client(addr, image string, width, height int) {
	t := thmb.NewThmb("unix", addr)
	defer t.Close()

	f, err := os.Create("thumbnail.jpg")
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	err = t.ResizeFile(image, f, uint32(width), uint32(height))
	if err != nil {
		log.Println(err)
		return
	}
}

func main() {
	var server, image string
	var width, height int
	var configFile string
	var chmod bool
	flag.StringVar(&configFile, "c", "/etc/thumbnailer.toml", "config file path")
	flag.StringVar(&server, "s", "", "server socket (setting this variable will activate client mode)")
	flag.StringVar(&image, "image", "", "path to image")
	flag.IntVar(&width, "w", 150, "width of the thumbnail")
	flag.IntVar(&height, "h", 150, "height of the thumbnail")
	flag.BoolVar(&chmod, "unsafe-perm", false, "chmod socket after creation to 777")
	flag.Parse()

	if server == "" && image == "" {
		serve(configFile, chmod)
	}

	if server != "" && image != "" {
		client(server, image, width, height)
	}

}
