package main

import (
  "fmt"
  "net/http"
  "code.google.com/p/go.net/html"
  "strings"
  "os"
  "io"
)

// Scrapes opensource.org for license texts

func getLicense(license string) {
  // Get the license's page
  resp, err := http.Get("http://opensource.org/licenses/"+license)
  if err != nil {
    fmt.Println(err)
    return
  }

  z := html.NewTokenizer(resp.Body)

  // state: 0 is active, -1 is searching for the container div
  state := -1

  // holds all the little textContent strings
  content := make([]string, 0, 64)

  // the nesting depth of the html tags, used to figure out when to pop out of
  // the div
  depth := 0

  // whether the loop should be running (break doesn't work because switch-case
  // I guess
  running := true
  for running {
    tok := z.Next()
    switch tok {
    case html.ErrorToken:
      running = false
    case html.StartTagToken:
      tagName, hasAttr := z.TagName()

      if state == -1 {
        if string(tagName) != "div" || !hasAttr {
          continue
        }
        for {
          attr, val, more := z.TagAttr()
          if string(attr) != "class" {
            if !more {
              break
            }
            continue
          }
          //fmt.Println("Found class: ", string(val))
          if string(val) != "field-item even" {
            if !more {
              break
            }
            continue
          }
          state = 0
        }
        continue
      }
      if state == 0 {
        depth ++
        //tagName, _ := z.TagName()
        //fmt.Println("increasing depth to ", depth, " with ", string(tagName))
      }

    case html.TextToken:
      if state == 0 {
        content = append(content, string(z.Text()))
      }
    case html.EndTagToken:
      if state == 0 {
        depth --
        //tagName, _ := z.TagName()
        //fmt.Println("decreasing depth to ", depth, " with ", string(tagName))
        if depth < 0 {
          running = false
        }
      }
    }
  }

  contentStr := strings.Join(content, "")
  file, err := os.Create(strings.ToLower(license))
  if err != nil {
    fmt.Println(err)
    return
  }
  io.WriteString(file, contentStr)
  file.Close()
}


func main() {
  urls := make([]string, 0, 75)

  resp, err := http.Get("http://opensource.org/licenses/alphabetical")
  if err != nil {
    fmt.Println(err)
    return
  }

  z := html.NewTokenizer(resp.Body)
  for {
    tok := z.Next()
    if tok == html.ErrorToken {
      //fmt.Println("reached error")
      break
    }
    if tok != html.StartTagToken {
      //fmt.Println("not a start tag")
      continue
    }

    tagName, hasAttr := z.TagName()
    if string(tagName) != "a" {
      //fmt.Println(string(tagName), " is not 'a'")
      continue
    }
    if !hasAttr {
      //fmt.Println("tag has no attributes")
      continue
    }

    href := ""

    for {
      attr, val, more := z.TagAttr()
      if string(attr) == "href" {
        //fmt.Println("Found href: ", string(val))
        href = string(val)
      }
      if !more {
        break
      }
    }
    if strings.HasPrefix(href, "/licenses/") {
      href = strings.Replace(href, "/licenses/", "", 1)
      if href == strings.ToLower(href) {
        continue
      }
      urls = append(urls, href)
    }
  }

  for _, license := range urls {
    getLicense(license)
  }
}
