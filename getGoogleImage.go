package main

import (
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"os"
	"github.com/labstack/echo/v4"
	"strings"
)

var source string
var globalCheck = false

// traverses until the first encounter of needed object of html DOM, first node of which is 'n'
func traverse(n *html.Node) {

	if globalCheck {
		return
	}
	pass := false
	for _, attr := range n.Attr {
		if attr.Val == "yWs4tf" { // name of a special class of a google search rresponse responsible for holding images
			pass = true
		}
		if attr.Key == "src" {
			if pass {
				source = attr.Val
				if source != "" {
					globalCheck = true
					return
				}
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		traverse(c)
	}
}

// takes url address of an Image and then saves it into a local .jpg file
func saveImage(src string) {

	response, e := http.Get(src)
	if e != nil {
		log.Fatal(e)
	}
	defer response.Body.Close()

	//opens a file for writing
	file, err := os.Create("yourImage.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}
}

func getImage(c echo.Context) error {

	nameObject := c.Param("id")

	fmt.Println(nameObject)
	// in case we receive several words as a one request
	names := strings.Split(nameObject, " ")
	nameObject = strings.Join(names, "+")
	// Getting the html response body, and writing it inside a newly created tmp file
	req, err := http.NewRequest("GET", "https://www.google.com/search?q=" + nameObject + "&hl=en&tbm=isch&source=hp&sclient=img", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// any filename will fit, as long as it is inside /tmp
	htmlFile, err := os.Create("/tmp/response.html")
	if err != nil {
		return err
	}
	defer htmlFile.Close()

	_, err = htmlFile.WriteString(string(b))
	if err != nil {
		return err
	}

	// at first, it seems that we could skip the var 'fin' and use 'htmlFile' instead, but
	// a file pointer created after os.Create() differs somewhere from a file pointer created after os.Open()
	fin, err := os.Open(htmlFile.Name())
	if err != nil {
		panic("Fail to open " + htmlFile.Name())
	}
	defer fin.Close()

	// representing html response file as DOM, and 'doc' will be its first node
	doc, err := html.Parse(fin)
	if err != nil {
		panic("Fail to parse " + htmlFile.Name())
	}

	traverse(doc)
	globalCheck = false // remove the check to traverse again on the next request
	saveImage(source)
	return c.String(http.StatusOK, "Request for " + nameObject + " went successful")
}

func main() {

	e := echo.New()
	e.GET("/getImage/:id", getImage)
	// http://localhost:1323/getImage/'yourObjectOfInterest'
	e.Logger.Fatal(e.Start(":1323"))
}