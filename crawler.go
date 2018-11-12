package main

import "net/http"
import "io/ioutil"
import "fmt"
import "regexp"
import "time"
import "encoding/json"
import "log"
import "gopkg.in/yaml.v2"
import "github.com/aws/aws-sdk-go/aws"
import "github.com/aws/aws-sdk-go/aws/session"
import "github.com/aws/aws-sdk-go/service/s3/s3manager"
import       "github.com/aws/aws-sdk-go/aws/credentials"
import         "github.com/aws/aws-sdk-go/service/s3"

import "bufio"
import "os"
//import "reflect"
//import "sort"

type coin struct {
    Url string  `json:"url"`
    Name string `json:"name"`
    Price string  `json:"price"`
};
type coinresult struct {
     status int
     coinpage [48]coin
};
type conf struct {
    AKey string `yaml:"accesskey"`
    SKey string `yaml:"secretkey"`
}


func (c *conf) getConf() *conf {

    yamlFile, err := ioutil.ReadFile("conf.yaml")
    if err != nil {
        log.Printf("yamlFile.Get err   #%v ", err)
    }
    err = yaml.Unmarshal(yamlFile, c)
    if err != nil {
        log.Fatalf("Unmarshal: %v", err)
    }

    return c
}



func fetch  (pagenum int, out chan<- coinresult) {

    var r coinresult;   

    client := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
        return http.ErrUseLastResponse
    } }

    baseurl:="https://www.numizmatik.ru/shopcoins/1894-1917-nikolai-ii_g574";
    if pagenum>1 {
        baseurl+=fmt.Sprintf("/?pagenum=%d",pagenum); 
       
    }

    resp, err := client.Get(baseurl)
    if err!=nil {
        panic (err);
    }
    defer resp.Body.Close();

    if resp.StatusCode==200 {
        body,err:=ioutil.ReadAll(resp.Body);
        if err!=nil {
            panic(err);
        }
        re    := regexp.MustCompile("(?s)<div class=\"good_name\">.+?<div class=\"good_price\">.+?</div>");   
        restr := regexp.MustCompile("(?s)<a itemprop=\"name\" href=\"(.+?)\".+?Монета 1894 – 1917 (.+?)</a>.+<meta itemprop=\"price\" content=\"(.+?)\">");
        i:=0;
        for _, value := range re.FindAllString(string(body), -1){
            //fmt.Println(value);
            m := restr.FindStringSubmatch(value) 
            if m!=nil {
                r.coinpage[i]=coin{
                    Url:m[1],
                    Name:m[2],
                    Price:m[3] };
                i++;
            }    
        }
        r.status=1;
    } else {    
        r.status=0;
    }
    out <- r;
    
}


func main () {
    start := time.Now();
    pages :=2;
    ch := make(chan coinresult);
    

    var c conf
    c.getConf()


    token := ""
    creds := credentials.NewStaticCredentials(c.AKey, c.SKey, token) 


    //aws.DefaultConfig.Region = "us-east-1" //<--- change this to yours

    conf := aws.Config{
        Region: aws.String("us-west-2"),
        Credentials:      creds,
    }

    sess := session.New(&conf)
    svc := s3manager.NewUploader(sess)

    fmt.Println("Uploading file to S3...")
    result, err := svc.Upload(&s3manager.UploadInput{
        Bucket: aws.String("vjpctest"),
        Key:    aws.String(filepath.Base(filename)),
        Body:   file,
    })
    if err != nil {
        fmt.Println("error", err)
        os.Exit(1)
    }
   

    region := aws.EUCentral   
          // change this to your AWS region
         // click on the bucketname in AWS control panel and click Properties
         // the region for your bucket should be under "Static Website Hosting" tab
    connection := s3.New(AWSAuth, region)
    bucket := connection.Bucket("vjpctest") // change this your bucket name
    path := "test.txt" // this is the target file and location in S3
    fileToBeUploaded := "crawler.go"
    file, err := os.Open(fileToBeUploaded)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer file.Close()

   fileInfo, _ := file.Stat()
   var size int64 = fileInfo.Size()
   bytes := make([]byte, size)

    // read into buffer
    buffer := bufio.NewReader(file)
    _, err = buffer.Read(bytes)

         // then we need to determine the file type
         // see https://www.socketloop.com/tutorials/golang-how-to-verify-uploaded-file-is-image-or-allowed-file-types

    filetype := http.DetectContentType(bytes)
    err = bucket.Put(path, bytes, filetype, s3.ACL("public-read"),s3.Options{})

    if err != nil {
        fmt.Println("put error",err)
        os.Exit(1)
    }

    fmt.Printf("Uploaded to %s with %v bytes to S3.\n\n", path, size)

  


    var cl []coin;
    cd := make(map[string]map[string][]string)
    for i:=1; i<pages; i ++ {
        go fetch(i,ch);
    }
    for i:=1; i<pages; i++ {
        
        cr:=<-ch;
        for i:=0; i<48; i++ {
            //fmt.Println(cr.coinpage[i].name);
            cl = append (cl, cr.coinpage[i]);
            re  := regexp.MustCompile(`Николай II (.+?) (\d{4})(.*)`); 
            m   := re.FindStringSubmatch(cr.coinpage[i].Name);
            if (m!=nil) {  
                if cd[m[1]] == nil { cd[m[1]] = make(map[string][]string) }    
               //fmt.Println(m[1],"----",m[2],"----",m[3]);
               cd[m[1]][m[2]]=append(cd[m[1]][m[2]],cr.coinpage[i].Price); 
            } else {
              fmt.Println("not found")  
            }  
        }
    }    
    for ck, cv := range cd {
       fmt.Println(ck); 
       fmt.Println(cv); 
              
    }  

    //fmt.Println(cd);


    //clist := reflect.ValueOf(cd).MapKeys()
    //sort.Strings(clist)
    //fmt.Println("-----",clist) 

    fmt.Println("=============")
    jsonStr, err := json.Marshal(cl)
    if (err!=nil) {
        panic(err)
    }
    fmt.Println(string(jsonStr))
   
    fmt.Printf("took %v\n",  time.Since(start))
  
}
