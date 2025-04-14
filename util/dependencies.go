package util

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"reflect"
)

func GetFieldNames(i interface{}) []string {
	var fields []string
	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fields
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		fields = append(fields, t.Field(i).Name)
	}
	return fields
}

func ExecCmd(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func ReadFileContent(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()
	contentBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return string(contentBytes), nil
}
