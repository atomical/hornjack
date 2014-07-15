package main

import (
  "os"
  "time"
  "flag"
  "fmt"
  "bytes"
  // "bufio"
  "io/ioutil"
  gq "github.com/PuerkitoBio/goquery"
  // "net/url"
  "net/http"
  "runtime"
  "strings"
  "net/url"
  "path"

)

const (
  BETA_EXPIRES_AT     = "Fri, 05 Sep 2014 15:04:05 MST"
  DEFAULT_USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/35.0.1916.153 Safari/537.36"
)

var inputURL string
var outputName string

func main(){
  setup()
  
  flag.StringVar(&inputURL, "u", "", "url for an avalon asset" )
  flag.StringVar(&outputName, "af", "", "filename for downloaded asset")

  flag.Parse()

  if (inputURL == "") {
    flag.PrintDefaults()
    return
  }
  
  html := fetchURL(inputURL)

  buf := bytes.NewBuffer(html)
  
  doc, err := gq.NewDocumentFromReader(buf)

  if err != nil {
    fmt.Println(err)
    return
  }
  
  tag := doc.Find("source[data-quality='high'][data-plugin-type='native']").First()
  src_url, _ := tag.Attr("src")
  playlist := fetchURL(src_url)

  var urls []string
  parsed_url, err := url.Parse(src_url)
  
  if err != nil {
    panic(err)
  }

  lines := strings.Split(string(playlist[:]), "\r\n")
  for i := range(lines) {
    if strings.HasSuffix(lines[i], ".ts") {
      downloadURL := fmt.Sprintf("http://%v%v/%v", parsed_url.Host, path.Dir(parsed_url.Path), lines[i])
      urls = append(urls, downloadURL)
    }
  }

  var mediaObject []byte
  for i := range(urls) {
    mediaFile := fetchURL(urls[i])
    mediaObject = append(mediaObject, mediaFile...)
  }

  if outputName == "" {
    outputName = path.Base(urls[0])
  }

  fmt.Println("Saving", outputName )

  ioutil.WriteFile(outputName, mediaObject, 0644)

}

func setup(){ 

  betaProtect()

  // utilize all cores
  cpus := runtime.NumCPU()
  runtime.GOMAXPROCS( cpus )

}

func betaProtect() {
  
  // usage indicator
  go func(){
    resp, _ := http.Get("http://www.adamhallett.com/hornjack")
    defer resp.Body.Close()
  }()

  // time limited
  t, _ := time.Parse(time.RFC1123, BETA_EXPIRES_AT)
  if time.Now().Unix() > t.Unix() {
    fmt.Println("Expired Beta Version")
    os.Exit(0)
  }

}

func fetchURL( url string ) ( []byte ) {


client := &http.Client{}

  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    fmt.Println(err)
    return nil
  }

  req.Header.Set("User-Agent", DEFAULT_USER_AGENT)

  resp, err := client.Do(req)
  if err != nil {
    fmt.Println(err)
    return nil
  }

  defer resp.Body.Close()
  
  body, err := ioutil.ReadAll(resp.Body)
  
  if err != nil {
    fmt.Println(err)
    return nil
  }

  return body 

}