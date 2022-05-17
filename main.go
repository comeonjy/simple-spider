// Package net1 @Description  TODO
// @Author  	 jiangyang
// @Created  	 2022/5/7 20:59
package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cobra"
)

type Spider struct {
	deep     int
	baseUrl  string
	savePath string
	index    []string
}

var root = cobra.Command{
	Use: "",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			log.Fatal("非法参数")
		}
		log.Println(args)
		if err := New(args[0], []string{"/index.html"}).Run(); err != nil {
			log.Println(err)
			return
		}
	},
}

func main() {
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func New(baseUrl string, urls []string) *Spider {
	return &Spider{
		deep:     3,
		baseUrl:  baseUrl,
		savePath: "./html",
		index:    urls,
	}
}

func (s *Spider) Run() error {
	urls := s.index
	for i := 0; i < s.deep; i++ {
		resUrls := make([]string, 0)
		for _, v := range urls {
			body, err := s.download(v)
			if err != nil {
				log.Println(err)
				continue
			}
			urls, err = s.fetch(body)
			if err != nil {
				log.Println(err)
				continue
			}
			resUrls = append(resUrls, urls...)
		}
		urls = resUrls
	}
	return nil
}

func (s *Spider) fetch(body []byte) ([]string, error) {
	var urls []string
	reg := regexp.MustCompile("^/.*")
	regDir := regexp.MustCompile("^/.*/$")

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	doc.Find("link").Each(func(i int, selection *goquery.Selection) {
		if v, ok := selection.Attr("href"); ok {
			if reg.MatchString(v) {
				if _, err := s.download(v); err != nil {
					log.Println(err)
					return
				}
			}
		}
	})
	doc.Find("img,script").Each(func(i int, selection *goquery.Selection) {
		if v, ok := selection.Attr("src"); ok {
			if reg.MatchString(v) {
				if _, err := s.download(v); err != nil {
					log.Println(err)
					return
				}
			}
		}
	})

	doc.Find("a").Each(func(i int, selection *goquery.Selection) {
		if v, ok := selection.Attr("href"); ok {
			if reg.MatchString(v) {
				if regDir.MatchString(v) {
					v += "index.html"
				}
				urls = append(urls, v)
			}
		}
	})
	return urls, nil
}

func (s *Spider) download(filename string) ([]byte, error) {
	if _, err := os.Stat(s.savePath + filename); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		return nil, nil
	}
	log.Println("GET ", s.baseUrl+filename)
	body, err := get(s.baseUrl + filename)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(s.savePath+filename), 0777); err != nil {
		return nil, err
	}

	if err := ioutil.WriteFile(s.savePath+filename, body, 0777); err != nil {
		return nil, err
	}
	return body, nil
}

func get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return body, nil
}
