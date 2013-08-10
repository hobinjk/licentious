package main

import (
  "encoding/json"
  "strings"
  "strconv"
  "path/filepath"
  "fmt"
  "net/http"
  "text/template"
  "os"
  "io/ioutil"
  "time"
)

type Licenser struct {
  templates map[string] *template.Template
}

func PathElems(path string) []string {
  parts := strings.Split(path, "/")
  elems := make([]string, 0, len(parts)-1)
  for _, part := range parts {
    if len(part) > 0 {
      elems = append(elems, part)
    }
  }
  return elems
}

type LicenseData struct {
  Holders string
  Year string
}

func (l* Licenser) ServeLicense(w http.ResponseWriter, r *http.Request) {
  elems := PathElems(r.URL.Path)
  // time.Year
  year := time.Now().Year()
  licenseData := &LicenseData{"<copyright holders>", strconv.Itoa(year)}
  values := r.URL.Query()
  for param, vals := range values {
    switch param {
    case "holder":
      switch len(vals) {
      case 1:
        licenseData.Holders = vals[0]
      case 2:
        licenseData.Holders = vals[0]+" and "+vals[1]
      default:
        licenseData.Holders = strings.Join(vals[:len(vals)-1], ", ")
        licenseData.Holders += ", and " + vals[len(vals)-1]
      }
    case "year":
      licenseData.Year = vals[0]
    }
  }


  if len(elems) > 1 {
    licenseData.Holders = elems[1]
  }
  licenseName := elems[0]
  if l.templates[licenseName] != nil {
    l.templates[licenseName].Execute(w, licenseData)
  } else {
    http.NotFound(w, r)
  }
}


func (l* Licenser) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  elems := PathElems(r.URL.Path)
  if len(elems) > 0 {
    l.ServeLicense(w, r)
    return
  }

  licenses := make([]string, 0, 16)
  for licenseName, _ := range l.templates {
    licenses = append(licenses, licenseName)
  }
  jsonData, err := json.Marshal(licenses)
  if err != nil {
    http.Error(w, "json failed to marshal", 500)
    fmt.Println("unable to marshal json")
    return
  }
  _, err = w.Write(jsonData)
  if err != nil {
    fmt.Println("unable to write json")
    return
  }
}

func main() {
  licenseDir, err := os.Open("licenses")
  if err != nil {
    panic(err)
  }

  fileInfos, err := licenseDir.Readdir(0)
  if err != nil {
    panic(err)
  }

  licenseDir.Close()

  lics := map[string] *template.Template {}

  for _, fileInfo := range fileInfos {
    file, err := os.Open(filepath.Join(licenseDir.Name(), fileInfo.Name()))
    if err != nil {
      fmt.Println("unable to read ", fileInfo.Name());
      continue
    }
    data, err := ioutil.ReadAll(file)
    if err != nil {
      fmt.Println("unable to read ", file.Name())
      continue
    }

    baseName := filepath.Base(file.Name())
    lics[baseName], err = template.New("").Parse(string(data))
    if err != nil {
      fmt.Println("unable to compile template ", file.Name(), " due to ", err)
      continue
    }
  }

  lic := &Licenser{lics}
  http.Handle("/licenses/", http.StripPrefix("/licenses/", lic));
  http.ListenAndServe(":31337", nil)
}
