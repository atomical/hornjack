package main

import (
  "os"
  "time"
  "flag"
  "fmt"
  "bytes"
  "io/ioutil"
  gq "github.com/PuerkitoBio/goquery"
  "net/http"
  "runtime"
  "strings"
  "net/url"
  "path"
  "github.com/jzelinskie/progress"
)

const (
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
  pb := progress.New(os.Stdin, int64(len(urls)))
  pb.Draw()
  pb.DrawEvery(1 * time.Second)

  for i := range(urls) {
    mediaFile := fetchURL(urls[i])
    mediaObject = append(mediaObject, mediaFile...)
    pb.Increment()
  }

  fmt.Println("\n")

  if outputName == "" {
    parsedInputURL, _ := url.Parse(inputURL)
    filename := strings.Split(path.Base(parsedInputURL.Path),":")[1] + ".ts"
    outputName = filename

  }

  ioutil.WriteFile(outputName, mediaObject, 0644)

}

func setup(){ 

  // utilize all cores
  cpus := runtime.NumCPU()
  runtime.GOMAXPROCS( cpus )

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