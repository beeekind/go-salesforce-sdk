// Package chromedp renders javascript web pages for the Salesforce
// documentation website using the chromedp/chromedp and PuerkitoBio/goquery
// libraries.
//
// Data scraped from the documentation pages is used to provide comments and
// docstrings for code generated in other packages.
//
// Example documentation pages:
// * https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_lead.htm
// * https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/tooling_api_objects_fielddefinition.htm
//
// Performance could be improved further by querying only parts of the page we care about such as
// https://developer.salesforce.com/docs/get_document_content/object_reference/sforce_api_objects_account.htm/en-us/230.0
// however I prefer the more publicly available root documentation page even if it is more resource intensive.
package chromedp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
)

type domain int

const (
	objects domain = iota
	tooling
	both

	toolingRoot = "https://developer.salesforce.com/docs/atlas.en-us.api_tooling.meta/api_tooling/reference_objects_list.htm"
	objectsRoot = "https://developer.salesforce.com/docs/atlas.en-us.object_reference.meta/object_reference/sforce_api_objects_list.htm"
)

var (
	// ErrNoHTMLTable ...
	ErrNoHTMLTable = errors.New("no data table found within htmml")
	// ErrNoDetailPageForRow ...
	ErrNoDetailPageForRow = errors.New("could not find link to documentation")
	// DefaultOptions configure chromedp to wait until specific page content is rendered before 
	// returning an html document. these configurations are specific to salesforce documentation
	// pages.
	DefaultOptions = []chromedp.Action{
		chromedp.WaitReady(":root"),
		chromedp.WaitVisible("h1.helpHead1"),
		//chromedp.WaitVisible(".section .data table tbody"),
		chromedp.ScrollIntoView(".prev-next-button"),
	}
)

// ParseWebApp is a high level API for retrieving a searchable HTML document from
// a JS-Rendered webpage
func ParseWebApp(url string, timeout time.Duration, options ...chromedp.Action) (*goquery.Document, error) {
	pageContents, err := GetWebApp(url, timeout, options...)
	if err != nil {
		return nil, fmt.Errorf("ParseWebApp(): %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(pageContents))
	if err != nil {
		return nil, fmt.Errorf("ParseWebApp(): goquery.NewDocumentFromReader(): %w", err)
	}

	return doc, nil
}

// contexts can be customized and reused if declared as globals here, however initial testing
// shows no meaningful improvements with these tweaks
// var ChromeCtx, _ = chromedp.NewContext(context.Background())
/**
var opts = append(
	chromedp.DefaultExecAllocatorOptions[:],
	chromedp.DisableGPU,
	chromedp.Flag("headless", false),
)
var rootCtx, _ = chromedp.NewExecAllocator(context.Background(), opts...)
var parentCtx, _ = chromedp.NewContext(rootCtx)
*/

// GetWebApp returns the results of a js-rendered web page
func GetWebApp(url string, timeout time.Duration, options ...chromedp.Action) (contents []byte, err error) {
	var outterHTML string

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	_, err = chromedp.RunResponse(ctx, chromedp.Tasks{
		// network.Enable(),
		chromedp.Navigate(url),
		// chromedp is inherently racey and may result in infinite blocks when trying to 
		// wait for query selectors to become rendered/visible. wrapping such options in a timeout
		// helps mitigate this problem. 
		runWithTimeout(&ctx, timeout, options),

		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}

			outterHTML, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	})

	if err != nil {
		return nil, fmt.Errorf("running response: %w", err)
	}

	return []byte(outterHTML), nil
}

func runWithTimeout(ctx *context.Context, timeout time.Duration, tasks chromedp.Tasks) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		timeoutContext, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return tasks.Do(timeoutContext)
	}
}

// GetObjects returns a map of objectName=>objectDetailPageURL for all standard objects
// for both the objects API and tooling.meta API
func GetObjects() (map[string]string, error) {
	return getObjects(both)
}

// GetStandardObjects returns a map of objectName=>objectDetailPageURL for all standard objects
func GetStandardObjects() (map[string]string, error) {
	return getObjects(objects)
}

// GetToolingObjects returns a map of objectName=>objectDetailPageURL for all tooling objects
func GetToolingObjects() (map[string]string, error) {
	return getObjects(tooling)
}

func getObjects(category domain) (map[string]string, error) {
	switch category {
	case objects:
		return getObject(objectsRoot)
	case tooling:
		return getObject(toolingRoot)
	case both:
		m1, err := getObject(objectsRoot)
		if err != nil {
			return nil, err
		}

		m2, err := getObject(toolingRoot)
		if err != nil {
			return nil, err
		}

		for k, v := range m2 {
			m1[k] = v
		}

		return m1, err

	default:
		return nil, fmt.Errorf("no documentation url for domain %v", category)
	}
}

func getObject(url string) (map[string]string, error) {
	doc, err := ParseWebApp(url, time.Second * 3, DefaultOptions...)
	if err != nil {
		return nil, err
	}

	results := map[string]string{}
	doc.Find(".ulchildlink").Each(func(idx int, s *goquery.Selection) {
		if err != nil {
			return
		}

		elem := s.Find("a").Last()
		if elem == nil || elem.Text() == "" || len(elem.Nodes) == 0 {
			err = ErrNoDetailPageForRow
			return
		}

		attr, exists := elem.Attr("href")
		if !exists || attr == "" {
			err = ErrNoDetailPageForRow
			return
		}

		if strings.Contains(attr, "https") {
			results[strings.TrimSpace(elem.Text())] = attr
		} else {
			results[strings.TrimSpace(elem.Text())] = "https://developer.salesforce.com/docs/" + attr
		}
	})

	return results, err
}