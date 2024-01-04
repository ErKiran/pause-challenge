package main

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"fb-down/insta"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	server()
}

func server() {
	http.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		url := query.Get("reel")

		res, err := insta.GetSnapInsta(url)
		if err != nil {
			fmt.Println("unale to get the insta post", err)
		}

		var resPath string

		if len(res.ResultMedia) != 0 {
			vidPath := downloadVideo(res.ResultMedia[0])

			if vidPath != "" {
				convertVideoToFrames(vidPath, "frames")
				imgMap := make(map[string]int)

				count := 0
				filepath.Walk("frames", func(path string, info os.FileInfo, err error) error {
					if path != "frames" {

						// todo image compare
						imgMap[path] = count
						count++
					}
					return nil
				})

				fmt.Println("imgMap", imgMap)

				for key, val := range imgMap {
					if len(imgMap) == 2 || len(imgMap) == 3 {
						if val == 1 {
							resPath = key
						}
					}
				}

				defer func() {
					removeAllFiles("video")
				}()
			}
		}

		// Set the content type to image/jpeg (or appropriate content type for your image).
		w.Header().Set("Content-Type", "image/jpeg")

		// Use http.ServeFile to send the image as the response.
		http.ServeFile(w, r, resPath)
	})

	// Start the server on port 8080.
	http.ListenAndServe(":8080", nil)
}

func removeAllFiles(root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fmt.Println("Removing file:", path)
			return os.Remove(path)
		}
		return nil
	})
}

func convertVideoToFrames(fileName, outputDir string) {
	cmd := exec.Command("ffmpeg", "-i", fileName, "-filter_complex", `select=bitor(gt(scene\,0.3)\,eq(n\,0))`, "-vsync", "drop", fmt.Sprintf("%s/%s", outputDir, "%04d.jpg"))
	var buffer bytes.Buffer
	cmd.Stdout = &buffer
	if cmd.Run() != nil {
		panic("could not generate frame")
	}
}

func downloadVideo(link string) string {
	filepath := "video"

	path := fmt.Sprintf("%s/%s.mp4", filepath, randSeq(10))

	// Create the file
	out, err := os.Create(path)
	if err != nil {
		return ""
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(link)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return ""
	}

	return path
}
