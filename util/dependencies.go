package util

import (
	"bytes"
	"fmt"
	"infiniband_exporter/log"
	"io"
	"os"
	"os/exec"
	"reflect"
	"slices"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	Cache     = make(map[string]map[string]string)
	CacheLock sync.RWMutex
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

func SetCache(configFilePath string) {
	yamlFile, err := os.ReadFile(configFilePath)
	if err != nil {
		log.GetLogger().Error(fmt.Sprintf("Read yaml file error:%s", err))
		os.Exit(1)
	}
	var linkMap map[string]map[string]string
	err = yaml.Unmarshal(yamlFile, &linkMap)
	if err != nil {
		log.GetLogger().Error("Unmarshal yaml file error")
		os.Exit(1)
	}

	CacheLock.Lock()
	for k, v := range linkMap {
		Cache[k] = v
	}
	CacheLock.Unlock()
}

func GetValueFromCache(key string) (map[string]string, bool) {
	CacheLock.RLock()
	defer CacheLock.RUnlock()
	val, exists := Cache[key]
	return val, exists
}
func GetKeysFromCache(guid string) bool {
	CacheLock.RLock()
	defer CacheLock.RUnlock()
	var guids []string
	for key := range Cache {
		keySplits := strings.Split(key, "_")
		if !slices.Contains(guids, keySplits[0]) {
			guids = append(guids, keySplits[0])
		}
	}
	return slices.Contains(guids, guid)
}
