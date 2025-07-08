package images

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/disintegration/imaging"
	log "github.com/sirupsen/logrus"
)

func horizontalImage(url string) *discordgo.File {
	resp, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return nil
	}
	defer resp.Body.Close()

	filename := strings.Split(url, "/")
	filename = strings.Split(filename[len(filename)-1], "?")
	format, err := imaging.FormatFromFilename(filename[0])
	if err != nil {
		log.Error(err)
		return nil
	}

	img, err := imaging.Decode(resp.Body)
	if err != nil {
		log.Error(err)
		return nil
	}

	img = imaging.Resize(img, img.Bounds().Dy()*15/10, img.Bounds().Dy(), imaging.Lanczos)

	if rand.Intn(128) == 0 {
		img = Fry(img)
	}

	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, img, format)
	if err != nil {
		log.Error(err)
		return nil
	}

	maxSize := 8 * 1024 * 1024
	if buf.Len() > maxSize {
		scaleFactor := float64(maxSize) / float64(buf.Len())
		newWidth := int(float64(img.Bounds().Dx()) * scaleFactor)
		newHeight := int(float64(img.Bounds().Dy()) * scaleFactor)
		img = imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
		buf.Reset()
		err = imaging.Encode(buf, img, format)
		if err != nil {
			log.Error(err)
			return nil
		}
	}

	return &discordgo.File{
		Name:        filename[0],
		ContentType: fmt.Sprintf("image/%s", format.String()),
		Reader:      buf,
	}
}

func horizontalVideo(url string) *discordgo.File {
	resp, err := http.Get(url)
	if err != nil {
		log.Error(err)
		return nil
	}
	defer resp.Body.Close()

	// Save the video to a temporary file
	tempFile, err := os.CreateTemp("", "video-*.mp4")
	if err != nil {
		log.Error(err)
		return nil
	}
	defer os.Remove(tempFile.Name())

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		log.Error(err)
		return nil
	}
	tempFile.Close()

	// Use ffmpeg to resize the video
	//	outputFile := strings.TrimSuffix(tempFile.Name(), ".mp4") + "-resized.mp4"
	//	cmd := exec.Command("ffmpeg", "-i", tempFile.Name(), "-c", "copy", "-metadata:s:v:0", "rotate=90", outputFile)
	cmd := exec.Command("exiftool", "-rotation<${rotation;$_ += 90}", tempFile.Name())
	// "ffmpeg -i input.mp4 -c copy -metadata:s:v:0 rotate=90 output.mp4"
	err = cmd.Run()
	if err != nil {
		log.Error(err)
		return nil
	}
	// Read the resized video into a buffer
	rotatedFile, err := os.Open(tempFile.Name())
	if err != nil {
		log.Error(err)
		return nil
	}
	defer rotatedFile.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, rotatedFile)
	if err != nil {
		log.Error(err)
		return nil
	}

	err = os.Remove(tempFile.Name())
	if err != nil {
		log.Error(err)
	}

	buf.ReadFrom(resp.Body)
	return &discordgo.File{
		Name:        "fixed.mp4",
		ContentType: "video/mp4",
		Reader:      buf,
	}
}
