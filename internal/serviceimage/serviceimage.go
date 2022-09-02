// Package serviceimage knows how to base64 encode a local image file
package serviceimage

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type ServiceImage string

// Encode will parse the ServiceImage value and replace with base64 encoded value if local file
func (s *ServiceImage) Encode(srcDir string) error {
	if strings.HasPrefix(string(*s), "file://") {
		imgPath := strings.Split(string(*s), "file://")
		encodedImage, err := base64EncodeImage(srcDir + imgPath[1])
		if err != nil {
			return err
		}
		*s = ServiceImage(encodedImage)
	}
	return nil
}

// base64EncodeImage returns a base64 encoded string for either jpg or png format files
func base64EncodeImage(filepath string) (string, error) {
	imgFile, err := os.ReadFile(filepath)

	if err != nil {
		return "", fmt.Errorf("unable to read service image file %s", filepath)
	}

	var encodedImage string
	contentType := http.DetectContentType(imgFile)

	switch contentType {
	case "image/jpeg":
		encodedImage += "data:image/jpeg;base64,"
	case "image/png":
		encodedImage += "data:image/png;base64,"
	}

	encodedImage += base64.StdEncoding.EncodeToString(imgFile)

	return encodedImage, nil
}
