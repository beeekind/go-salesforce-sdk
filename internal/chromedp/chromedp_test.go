package chromedp_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/beeekind/go-salesforce-sdk/internal/async"
	"github.com/beeekind/go-salesforce-sdk/internal/chromedp"
	"github.com/stretchr/testify/require"
)

const tmpl = "https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_%s.htm"

var pool = async.New(30, nil)

type workResult struct {
	Endpoint string
	Err      error
}

func work(subdir string, objectName string, docURL string) (string, error) {
	doc, err := chromedp.ParseWebApp(docURL, time.Second*3, chromedp.DefaultOptions...)
	if err != nil {
		return docURL, err
	}
	errMsgHidden := doc.Find(".error-container.ng-hide h2").Text()
	errMsg := doc.Find(".error-container h2").Text()
	// tableSelection := doc.Find("table tbody").First()
	entityName := doc.Find("h1.helpHead1").Text()
	entityDoc := doc.Find(".shortdesc").Text()
	table := doc.Find(".section .data table tbody").First()

	if errMsgHidden == "" && errMsg != "" {
		return docURL, fmt.Errorf("%s: is 404: %s\n", errMsg, docURL)
	}

	if entityName == "" || entityDoc == "" {
		return docURL, fmt.Errorf("is empty: %s\n", docURL)
	}

	if len(table.Nodes) == 0 {
		// return objectName, chromedp.ErrNoHTMLTable
	}

	index := 0
	table.Find("tr").Each(func(idx int, s *goquery.Selection) {
		index++
	})

	html, _ := doc.Html()
	filePath := fmt.Sprintf("./%s/%s.html", subdir, objectName)
	return docURL, writeFile(filePath, 0755, []byte(html))
}

func TestGetStandardObjects(t *testing.T) {
	urls, err := chromedp.GetStandardObjects()
	require.Nil(t, err)
	require.Greater(t, len(urls), 0)
}

func TestGetToolingObjects(t *testing.T) {
	urls, err := chromedp.GetToolingObjects()
	require.Nil(t, err)
	require.Greater(t, len(urls), 0)
}

func TestParseWebApp(t *testing.T) {
	docs, err := chromedp.GetObjects()
	require.Nil(t, err)

	var finalObjects = map[string]string{}
	for objectName, docURL := range docs {
		if _, err := os.Stat("./all/" + objectName + ".html"); os.IsNotExist(err) {
			// ignore xml, json, and other file extensions
			if strings.HasSuffix(docURL, ".htm") {
				finalObjects[objectName] = docURL
			}

		}
	}

	var inputs []async.Closure
	var outputs []interface{}
	idx := 0
	for k, v := range finalObjects {
		outputs = append(outputs, &workResult{})
		inputs = append(inputs, async.MustInput(0, work, outputs[idx], "all", k, v))
		idx++
	}

	if _, err := pool.Retry(inputs, time.Second*5, time.Second*5, time.Second*5); err != nil {
		println(-1, "failed:", err.Error())
	}

	for idx, result := range outputs {
		item := result.(*workResult)
		if item.Err != nil {
			println(idx, item.Endpoint, item.Err.Error())
		} else {
			println(idx, item.Endpoint)
		}
	}
}

// writeFile is a utility method for creating a file on a fully qualified path
// writeFile will make any subdirectories not already present on the absolutePath
// writeFile will overwrite existing files on the given absolute path
func writeFile(absolutePath string, perm os.FileMode, contents []byte) error {
	if err := os.MkdirAll(filepath.Dir(absolutePath), perm); err != nil {
		return fmt.Errorf("WriteFile(): failed MkdirAll: %w", err)
	}

	file, err := os.OpenFile(absolutePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("WriteFile(): failed OpenFile: %w", err)
	}

	defer file.Close()
	if _, err := file.Write(contents); err != nil {
		return fmt.Errorf("WriteFile(): failed file.Write: %w", err)
	}

	return nil
}

func stripNewLinesAndTabs(str string) string {
	str = strings.ReplaceAll(str, "\n", " ")
	str = strings.ReplaceAll(str, "\t", "")
	return strings.Join(strings.Fields(str), " ")
}
