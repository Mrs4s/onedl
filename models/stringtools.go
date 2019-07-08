package models

import (
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func GetParentPath(path string) string {
	list := strings.Split(path, "/")
	var tmp []string
	for i := 0; i < len(list)-1; i++ {
		tmp = append(tmp, list[i])
	}
	parentPath := strings.Join(tmp, "/")
	if parentPath == "" {
		return "/"
	}
	return parentPath
}

func GetFileName(path string) string {
	length := len(path)
	index := length - 1
	for index > 0 {
		char := path[index-1 : index]
		if char == "\\" || char == "/" || char == ":" {
			return path[index:length]
		}
		index--
	}
	return path
}

func GetSizeString(size int64) string {
	switch {
	case size == 0:
		return "nil"
	case size < 1024:
		return strconv.FormatInt(size, 10) + "B"
	case size < 1024*1024:
		return strconv.FormatInt(size/1024, 10) + "KB"
	default:
		return strconv.FormatInt(size/1024/1024, 10) + "MB"
	}
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func CombinePaths(path1 string, path2 string) string {
	if len(path2) == 0 {
		return path1
	}
	if len(path1) == 0 {
		return path2
	}
	char := path1[len(path1)-1:]
	if char != "\\" && char != "/" && char != ":" {
		return path1 + string(os.PathSeparator) + path2
	}
	return path1 + path2
}

func FilterChinese(text string) string {
	if strings.Contains(runtime.GOOS, "windows") {
		reg := regexp.MustCompile("[^\u4e00-\u9fa5]")
		res := strings.Join(reg.FindAllString(text, -1), "")
		for _, word := range []string{"（", "）", "【", "】", "：", "。"} {
			res = strings.ReplaceAll(res, word, "")
		}
		return res

	}
	return text
}
