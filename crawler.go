package main

import "net/http"
import "io/ioutil"
import "fmt"
import "regexp"
import "time"
import "encoding/json"



func fetch  (pagenum int, out chan<- string) {

    cns := make(map[string]interface{});

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
        re    := regexp.MustCompile("(?s)<span class=\"product__name\" itemprop=\"name\">.+?<div class=\"product__price\">.+?</div>");   
        restr := regexp.MustCompile("(?s).+?<a href='(.+?)'.+?Монета 1894 – 1917 (.+?)</a>.+<meta itemprop=\"price\" content=\"(.+?)\">");
        for _, value := range re.FindAllString(string(body), -1){
            m := restr.FindStringSubmatch(value) 
            if m!=nil {
                curl:=m[1];
                cns[curl]=map[string]interface{}{
                        "url":m[1],
                        "name":m[2],
                        "price":m[3] };
            }    
        }
        j, err := json.Marshal(cns)
        if err!=nil {
            panic(err);
        }
        out <- string(j);
    } else {    
        out <- "{}";
    }    
}


func main () {
    start := time.Now();
    pages :=10;
    

    ch := make(chan string)

    for i:=1; i<pages; i ++ {
        go fetch(i,ch);
    }
    for i:=1; i<pages; i++ {
        
        j:=<-ch;

        var data interface{};
        err := json.Unmarshal([]byte(j), &data);

        if err!=nil {
            panic(err);
        }

        fmt.Println(data);
    }    
    fmt.Printf("took %v\n",  time.Since(start))
  
}
