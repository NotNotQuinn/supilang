package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

// https://github.com/Supinic/supibot-package-manager/blob/master/commands/pastebin/index.js#L14
var (
	allowedGistTypes = []string{"text/plain", "application/javascript"}
	cacheDir         = "./.gistcache/"
)

type githubGistAPIResp struct {
	Files     map[string]gistFile `json:"files"`
	Truncated bool                `json:"truncated"`
	Message   string              `json:"message"`
}

type gistFile struct {
	Type      string `json:"type"`
	Truncated bool   `json:"truncated"`
	Content   string `json:"content"`
}

func getGistContent(id string) (string, error) {
	if match, err := regexp.MatchString("^[0-9a-fA-F]*$", id); !match || err != nil {
		if err != nil {
			return "", err
		}
		return "", errors.New("gist ids can only contain hexadecimal characters (0123456789abcdefABCDEF)")
	}
	if id == "" {
		return "", errors.New("a gist id cannot be the empty string")
	}

	cacheFile, err := filepath.Abs(filepath.Join(cacheDir, id))
	if err != nil {
		return "", err
	}
	cachedContent, err := os.ReadFile(cacheFile)
	if errors.Is(err, fs.ErrNotExist) {
		resp, err := http.Get("https://api.github.com/gists/" + id)
		if err != nil {
			return "", err
		}
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		err = resp.Body.Close()
		if err != nil {
			return "", err
		}
		respJson := &githubGistAPIResp{}
		err = json.Unmarshal(bytes, respJson)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != 200 {
			return "", errors.New(respJson.Message)
		}
		if len(respJson.Files) == 0 {
			return "", errors.New("there are no files in this Gist")
		}
		eligibleFiles := []gistFile{}
		for _, v := range respJson.Files {
			for _, allowedType := range allowedGistTypes {
				if v.Type == allowedType {
					eligibleFiles = append(eligibleFiles, v)
					break
				}
			}
		}
		if len(eligibleFiles) == 0 {
			return "", errors.New("no eligible files found in this Gist")
		}
		if len(eligibleFiles) > 1 {
			return "", errors.New("too many eligible files found in this Gist")
		}

		content := eligibleFiles[0].Content

		err = os.MkdirAll(cacheDir, 0644)
		if err != nil {
			return "", err
		}

		err = os.WriteFile(cacheFile, []byte(content), 0644)
		if err != nil {
			return "", err
		}
		return content, nil
	} else if err != nil {
		return "", err
	} else {
		return string(cachedContent), nil
	}
}
