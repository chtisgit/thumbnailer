package main

import (
	"flag"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/chtisgit/thumbnailer/thmb"
)

func serve(configFile string) {
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

	log.Println("Ready.")
	err = srv.Serve(nil)
	log.Print(err)
}

func client(addr, image string, width, height int) {
	t := thmb.NewThmb("unix", addr)
	defer t.Close()

	b, err := t.ResizeFile(image, uint32(width), uint32(height))
	if err != nil {
		log.Println(err)
		return
	}

	f, err := os.Create("thumbnail.jpg")
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	f.Write(b)
}

func main() {
	var server, image string
	var width, height int
	var configFile string
	flag.StringVar(&configFile, "c", "/etc/thumbnailer.toml", "config file path")
	flag.StringVar(&server, "s", "", "server socket (setting this variable will activate client mode)")
	flag.StringVar(&image, "image", "", "path to image")
	flag.IntVar(&width, "w", 150, "width of the thumbnail")
	flag.IntVar(&height, "h", 150, "height of the thumbnail")
	flag.Parse()

	if server == "" && image == "" {
		serve(configFile)
	}

	if server != "" && image != "" {
		client(server, image, width, height)
	}

}
