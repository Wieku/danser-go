package utils

import (
	"archive/zip"
	"fmt"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/texture"
	_ "golang.org/x/image/bmp"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getPixmap(name string) (*texture.Pixmap, error) {
	if strings.HasPrefix(name, "assets") {
		return assets.GetPixmap(name)
	}

	return texture.NewPixmapFileString(name)
}

func LoadTexture(path string) (*texture.TextureSingle, error) {
	log.Println("Loading texture:", path)

	img, err := getPixmap(path)

	if err == nil {
		defer img.Dispose()

		tex := texture.NewTextureSingle(img.Width, img.Height, 0)
		tex.Bind(0)
		tex.SetData(0, 0, img.Width, img.Height, img.Data)

		return tex, nil
	}

	log.Println("Failed to read a texture: ", err)

	return nil, err
}

func LoadTextureToAtlas(atlas *texture.TextureAtlas, path string) (*texture.TextureRegion, error) {
	log.Println("Loading texture into atlas:", path)

	img, err := getPixmap(path)

	if err == nil {
		defer img.Dispose()
		return atlas.AddTexture(path, img.Width, img.Height, img.Data), nil
	}

	log.Println("Failed to read a texture: ", err)

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
