package main

import "net/http"
import "io/ioutil"
import "fmt"
import "regexp"
import "time"


type coin struct{
    name string
    price string
};


func fetch  (pagenum int, out chan<- string) {
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
        if resp.StatusCode==200 {
            defer resp.Body.Close();
            body,err:=ioutil.ReadAll(resp.Body);
            if err!=nil {
                panic(err);
            }
            re    := regexp.MustCompile("(?s)<span class=\"product__name\" itemprop=\"name\">.+?<div class=\"product__price\">.+?</div>");   
            restr := regexp.MustCompile("(?s)Монета 1894 – 1917 (.+?)</a>.+<meta itemprop=\"price\" content=\"(.+?)\">");
            for _, value := range re.FindAllString(string(body), -1){
                m := restr.FindStringSubmatch(value) 
                if m!=nil {
                    c := coin {name:m[1],price:m[2]};
                    fmt.Println(c.name,"----",c.price);
                }    
            }
            out <- fmt.Sprintf("num:%d success %s",pagenum,baseurl);
        } else {    
            out <- fmt.Sprintf("num:%d skip %s",pagenum,baseurl);
        }    
}


func main () {
    start := time.Now();
    pages :=10;

    ch := make(chan string)

    for i:=1; i<pages; i ++ {
        go fetch(i,ch);
    }
    for j:=1; j<pages; j++ {
        fmt.Println(<-ch);
    }    
    fmt.Printf("took %v\n",  time.Since(start))
  
}
