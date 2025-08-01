package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime/debug"
)

type Flag struct {
	version bool
	verbose bool

	projectID  string
	apiVersion string
	token      string
	file       string

	targetEndpointFolderId        int
	targetSchemaFolderId          int
	endpointOverwriteBehavior     string
	schemaOverwriteBehavior       string
	updateFolderOfChangedEndpoint bool
	prependBasePath               bool
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	var f Flag
	flag.BoolVar(&f.version, "v", false, "show version")
	flag.BoolVar(&f.verbose, "verbose", false, "more output")

	flag.StringVar(&f.projectID, "projectID", "", "apifox project id")
	flag.StringVar(&f.apiVersion, "apiver", "2024-03-28", "the value of http header X-Apifox-Api-Version")
	flag.StringVar(&f.token, "token", "", "apifox api token")
	flag.StringVar(&f.file, "file", "swagger.yaml", "swagger file path")

	flag.IntVar(&f.targetEndpointFolderId, "targetEndpointFolderId", 0, "REF https://apifox-openapi.apifox.cn/api-173409873")
	flag.IntVar(&f.targetSchemaFolderId, "targetSchemaFolderId", 0, "REF https://apifox-openapi.apifox.cn/api-173409873")
	flag.StringVar(&f.endpointOverwriteBehavior, "endpointOverwriteBehavior", "OVERWRITE_EXISTING", "REF https://apifox-openapi.apifox.cn/api-173409873")
	flag.StringVar(&f.schemaOverwriteBehavior, "schemaOverwriteBehavior", "OVERWRITE_EXISTING", "REF https://apifox-openapi.apifox.cn/api-173409873")
	flag.BoolVar(&f.updateFolderOfChangedEndpoint, "updateFolderOfChangedEndpoint", false, "REF https://apifox-openapi.apifox.cn/api-173409873")
	flag.BoolVar(&f.prependBasePath, "prependBasePath", false, "REF https://apifox-openapi.apifox.cn/api-173409873")

	flag.Parse()

	if f.version {
		_, mainVer := getBuildVersion()
		fmt.Println("git_version:", mainVer)
		fmt.Println("build_time:", BuildTime)
		return
	}

	input, err := read(f.file)
	if err != nil {
		log.Fatal(err)
	}

	if err := request(input, &f); err != nil {
		log.Fatal(err)
	}

	fmt.Println("apifox import completed.")
}

// https://apifox-openapi.apifox.cn/api-173409873
//
//	curl -L 'https://api.apifox.com/v1/projects/12345/import-openapi?locale=zh-CN' \
//	-H 'X-Apifox-Api-Version: 2024-03-28' \
//	-H 'Authorization: Bearer xxx' \
//	-H 'Content-Type: application/json' \
//	-d '{
//	    "input": "xxx",
//	    "options": {
//	        "targetEndpointFolderId": 0,
//	        "targetSchemaFolderId": 0,
//	        "endpointOverwriteBehavior": "OVERWRITE_EXISTING",
//	        "schemaOverwriteBehavior": "OVERWRITE_EXISTING",
//	        "updateFolderOfChangedEndpoint": true,
//	        "prependBasePath": true
//	    }
//	}'
func request(input string, f *Flag) error {
	type Options struct {
		TargetEndpointFolderID        int    `json:"targetEndpointFolderId"`
		TargetSchemaFolderID          int    `json:"targetSchemaFolderId"`
		EndpointOverwriteBehavior     string `json:"endpointOverwriteBehavior"`
		SchemaOverwriteBehavior       string `json:"schemaOverwriteBehavior"`
		UpdateFolderOfChangedEndpoint bool   `json:"updateFolderOfChangedEndpoint"`
		PrependBasePath               bool   `json:"prependBasePath"`
	}
	type Payload struct {
		Input   string  `json:"input"`
		Options Options `json:"options"`
	}

	data := Payload{
		Input: input,
		Options: Options{
			TargetEndpointFolderID:        f.targetEndpointFolderId,
			TargetSchemaFolderID:          f.targetSchemaFolderId,
			EndpointOverwriteBehavior:     f.endpointOverwriteBehavior,
			SchemaOverwriteBehavior:       f.schemaOverwriteBehavior,
			UpdateFolderOfChangedEndpoint: f.updateFolderOfChangedEndpoint,
			PrependBasePath:               f.prependBasePath,
		},
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("encode payload: %w", err)
	}
	body := bytes.NewReader(payloadBytes)

	api := fmt.Sprintf("https://api.apifox.com/v1/projects/%s/import-openapi?locale=zh-CN", f.projectID)
	req, err := http.NewRequest(http.MethodPost, api, body)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Apifox-Api-Version", f.apiVersion)
	req.Header.Set("Authorization", "Bearer "+f.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request import: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if f.verbose {
		fmt.Printf("Result: %s\n", b)
	}

	return nil
}

func read(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read swagger file (%s): %w", path, err)
	}
	return string(b), nil
}

var BuildTime = "unknown"

func getBuildVersion() (goVersion, mainVersion string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown", "unknown"
	}
	return info.GoVersion, info.Main.Version
}
