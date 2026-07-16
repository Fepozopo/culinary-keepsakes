package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Fepozopo/culinary-keepsakes/src/website"
)

// main generates metadata-driven content pages, renders the static site, and serves the docs directory locally.
// Passing --build writes generated content and docs without starting the local HTTP server.
func main() {
	// A build-only mode lets build.sh create committed site artifacts without depending on a running server.
	buildOnly := len(os.Args) > 1 && os.Args[1] == "--build"

	// Grab the basepath from the first CLI argument or default to "/".
	basepath := "/"
	if len(os.Args) > 1 && !buildOnly {
		basepath = os.Args[1]
		if !strings.HasSuffix(basepath, "/") {
			basepath += "/"
		}
		if !strings.HasPrefix(basepath, "/") {
			basepath = "/" + basepath
		}
	}

	// Generate listing pages before rendering so content remains the source used to create docs.
	err := website.GenerateContentIndexes("content")
	if err != nil {
		fmt.Println("Error generating content indexes:", err)
		return
	}
	fmt.Println("Content indexes generated successfully.")

	// Copy static files to docs directory.
	err = website.CopyStaticToPublic()
	if err != nil {
		fmt.Println("Error copying static files:", err)
		return
	}
	fmt.Println("Static files copied successfully.")

	// Generate the markdown pages
	err = website.GeneratePagesRecursive("content", "template.html", "docs", basepath)
	if err != nil {
		fmt.Println("Error generating pages:", err)
		return
	}
	fmt.Println("Pages generated successfully.")
	if buildOnly {
		return
	}

	// Start the HTTP server to serve the docs directory, including static files with basepath.
	fs := http.FileServer(http.Dir("docs"))
	http.Handle(basepath, http.StripPrefix(basepath, fs))

	fmt.Println("Server started at http://localhost:8080" + basepath)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
