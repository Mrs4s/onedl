package onedrive

import (
	"github.com/thedevsaddam/gojsonq"
	"net/url"
	"onedl/models"
	"regexp"
	"strconv"
	"strings"
)

type SharePointShared struct {
	Url string

	requestUrl string
	listUrl    string
	viewId     string
}

func (shared *SharePointShared) Init() bool {
	page := models.GetStrings(shared.Url)
	if page == "" {
		return false
	}
	reg := regexp.MustCompile("_spPageContextInfo=(.*?)};")
	contextJson := strings.ReplaceAll(reg.FindString(page), "_spPageContextInfo=", "")
	contextJson = contextJson[0 : len(contextJson)-1]
	json := gojsonq.New().JSONString(contextJson)
	if json.Error() != nil {
		return false
	}
	shared.requestUrl = json.Find("siteAbsoluteUrl").(string)
	shared.listUrl = json.Reset().Find("listUrl").(string)
	shared.viewId = json.Reset().Find("viewId").(string)
	shared.viewId = shared.viewId[1 : len(shared.viewId)-2]
	return true
}

func (shared *SharePointShared) GetRootFiles() []SharedFile {
	return shared.GetFilesByPath("/")
}

func (shared *SharePointShared) GetFilesByPath(path string) []SharedFile {
	if path == "" {
		path = "/"
	}
	var result []SharedFile
	listBody := `{"parameters": {"__metadata": {"type": "SP.RenderListDataParameters"}, "RenderOptions": 12551, "AllowMultipleValueFilterForTaxonomyFields": true, "AddRequiredFields": true}}`
	listJson := gojsonq.New().JSONString(models.PostJson(shared.requestUrl+"/_api/web/GetList(@listUrl)/RenderListDataAsStream?@listUrl='"+shared.listUrl+"'&View="+shared.viewId+"&RootFolder="+shared.listUrl+"/path"+url.PathEscape(path), listBody))
	if listJson.Error() != nil {
		return nil
	}
	for _, file := range filterSharedFile(listJson.Reset().Find("ListData.Row").([]interface{}), path, shared) {
		result = append(result, file)
	}
	if next := listJson.Reset().Find("ListData.NextHref"); next != nil {
		for {
			listJson = gojsonq.New().JSONString(models.PostJson(shared.requestUrl+"/_api/web/GetList(@listUrl)/RenderListDataAsStream"+next.(string)+"&@listUrl='"+shared.listUrl+"'", listBody))
			for _, file := range filterSharedFile(listJson.Find("ListData.Row").([]interface{}), path, shared) {
				result = append(result, file)
			}
			if next = listJson.Reset().Find("ListData.NextHref"); next == nil {
				break
			}
		}
	}
	return result
}

func filterSharedFile(rows []interface{}, path string, shared *SharePointShared) []SharedFile {
	var result []SharedFile
	for _, row := range rows {
		obj := row.(map[string]interface{})
		name := obj["FileLeafRef"].(string)
		var realPath string
		if path[len(path)-1:] == "/" {
			realPath = path + name
		} else {
			realPath = path + "/" + name
		}
		sizeDisplay, ok := obj["FileSizeDisplay"]
		var size int64
		if ok {
			size, _ = strconv.ParseInt(sizeDisplay.(string), 10, 64)
		} else {
			size = 0
		}
		file := SharedFile{
			Owner:     shared,
			Name:      name,
			Path:      realPath,
			IsDir:     obj["FSObjType"].(string) == "1",
			Size:      size,
			SpItemUrl: obj[".spItemUrl"].(string),
			GetDownloadLink: func(file SharedFile) string {
				if file.IsDir {
					return ""
				}
				json := gojsonq.New(gojsonq.SetSeparator("->")).JSONString(models.GetStrings(file.SpItemUrl))
				return json.Find("@content.downloadUrl").(string)
			},
		}
		result = append(result, file)
	}
	return result
}
