package env

import (
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"strings"
)

var cache = make(map[string]string)

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

func Get(key string) string {
	val, exists := cache[key]
	if exists {
		return val
	}

	filename := viper.GetString(key + ".file")
	if filename == "" {
		return viper.GetString(key)
	}
	val, err := readSecret(filename)
	if err != nil {
		log.Printf("error reading secret: %s", err.Error())
	}
	//update cache with the full value, so we don't constantly read it
	cache[key] = val
	return val
}

func Set(key string, val string) {
	cache[key] = val
}

func GetOr(key string, def string) string {
	res := Get(key)
	if res == "" {
		return def
	}
	return res
}

func GetBool(key string) bool {
	return GetBoolOr(key, false)
}

func GetBoolOr(key string, def bool) bool {
	res := Get(key)
	if res == "" {
		return def
	}
	return cast.ToBool(res)
}

func GetInt(key string) int {
	return cast.ToInt(Get(key))
}

func GetStringArray(key, separator string) []string {
	val := Get(key)
	if separator == "" {
		separator = ","
	}
	return strings.Split(val, separator)
}

func readSecret(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(data)), nil
}
