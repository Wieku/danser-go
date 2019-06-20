package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	_ "image/jpeg"
	_ "image/gif"
	_ "golang.org/x/image/bmp"
	_ "image/png"
	"image"
	"image/draw"
	"log"
	"github.com/wieku/danser-go/render/texture"
	"path/filepath"
	"strings"
)

func LoadImage(path string) (*image.NRGBA, error) {
	file, err := os.Open(path)
	log.Println("Loading texture: ", path)
	if err != nil {
		log.Println("er1")
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		log.Println("er2")
		return nil, err
	}
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, bounds.Min, draw.Src)
	return nrgba, nil
}

/*func LoadTexture(path string) (*texture.Texture, error) {
	img, err := LoadImage(path)
	if err == nil {
		tex := glhf.NewTexture(
			img.Bounds().Dx(),
			img.Bounds().Dy(),
			4,
			true,
			img.Pix,
		)

		tex.Begin()
		tex.SetWrap(glhf.CLAMP_TO_EDGE)
		tex.End()

		return tex, nil
	}
	return nil, err
}*/

func LoadTexture(path string) (*texture.TextureSingle, error) {
	img, err := LoadImage(path)
	if err == nil {
		tex := texture.LoadTextureSingle(img, 4)

		return tex, nil
	}
	return nil, err
}

func LoadTextureToAtlas(atlas *texture.TextureAtlas, path string) (*texture.TextureRegion, error) {
	img, err := LoadImage(path)
	if err == nil {
		return atlas.AddTexture(path, img.Bounds().Dx(), img.Bounds().Dy(), img.Pix), nil
	}
	log.Println(err)
	return nil, err
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}

/*func LoadTextureU(path string) (*glhf.Texture, error) {
	img, err := LoadImage(path)
	if err == nil {
		tex := glhf.NewTexture(
			img.Bounds().Dx(),
			img.Bounds().Dy(),
			0,
			true,
			img.Pix,
		)

		tex.Begin()
		tex.SetWrap(glhf.CLAMP_TO_EDGE)
		tex.End()

		return tex, nil
	}
	return nil, err
}*/
