package helpers

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/jfrog/jfrog-client-go/utils/io/fileutils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"
)

func ResolvePathAndPanicIfNotFound(path string) string {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	if !fileutils.IsPathExists(absolutePath, true) {
		panic(fmt.Errorf("File not found: %s", absolutePath))
	}
	return absolutePath
}

func DownloadFile(url string, targetPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func ExtractTarGz(gzipStream io.Reader, target string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}
	tarReader := tar.NewReader(uncompressedStream)
	var header *tar.Header
	for header, err = tarReader.Next(); err == nil; header, err = tarReader.Next() {
		targetName := filepath.Join(target, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(targetName, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.OpenFile(targetName, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			if err := outFile.Close(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("uknown type: %b in %s", header.Typeflag, header.Name)
		}
	}
	if err != io.EOF {
		return err
	}
	return nil
}

func GetStructAsEnvironmentVariables(object interface{}) map[string]string {
	environmentVariables := make(map[string]string)
	properties := reflect.ValueOf(object)
	for i := 0; i < properties.Type().NumField(); i++ {
		propertyName := properties.Type().Field(i).Name
		propertyValue := properties.FieldByName(propertyName)
		//if !propertyValue.IsZero() {
		key := ConvertFirstRuneToLowerCase(propertyName)
		switch propertyValue.Kind() {
		case reflect.Slice:
			lengthKey := fmt.Sprintf("%s_len", key)
			environmentVariables[lengthKey] = fmt.Sprintf("%v", propertyValue.Len())
			for j := 0; j < propertyValue.Len(); j++ {
				sliceItem := propertyValue.Index(j)
				sliceItemKey := fmt.Sprintf("%s_%v", key, j)
				environmentVariables[sliceItemKey] = GetEnvironmentVariableFriendlyValue(sliceItem.Interface())
			}
		case reflect.Array:
			lengthKey := fmt.Sprintf("%s_len", key)
			environmentVariables[lengthKey] = fmt.Sprintf("%v", propertyValue.Type().Len())
			for j := 0; j < propertyValue.Len(); j++ {
				sliceItem := propertyValue.Index(j)
				sliceItemKey := fmt.Sprintf("%s_%v", key, j)
				environmentVariables[sliceItemKey] = GetEnvironmentVariableFriendlyValue(sliceItem.Interface())
			}
		}
		environmentVariables[key] = GetEnvironmentVariableFriendlyValue(propertyValue.Interface())
		//}
	}
	return environmentVariables
}

func GetEnvironmentVariableFriendlyValue(value interface{}) string {
	if value == nil {
		return ""
	}
	jsonValue, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Errorf("failed to get integration value and environment variable: %w", err))
	}
	stringValue := string(jsonValue)
	stringValue = strings.Trim(stringValue, "\"")
	stringValue = strings.ReplaceAll(stringValue, "\"", "\\\"")
	return stringValue
}

func ConvertFirstRuneToLowerCase(name string) string {
	nameRunes := []rune(name)
	nameRunes[0] = unicode.ToLower(nameRunes[0])
	return string(nameRunes)
}
