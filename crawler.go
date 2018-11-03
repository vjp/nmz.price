package main

import "net/http"
import "io/ioutil"
import "fmt"
import "regexp"
import "time"
//import "container/list"

type coin struct {
    url string
    name string
    price string
};
type coinresult struct {
     status int
     coinpage [48]coin
};


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
                    url:m[1],
                    name:m[2],
                    price:m[3] };
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
    pages :=20;
    
    ch := make(chan coinresult);
    
    var cl []coin;
    cd := make(map[string][]string) 

    for i:=1; i<pages; i ++ {
        go fetch(i,ch);
    }
    for i:=1; i<pages; i++ {
        
        cr:=<-ch;
        for i:=0; i<48; i++ {
            //fmt.Println(cr.coinpage[i].name);
            cl = append (cl, cr.coinpage[i]);
            re  := regexp.MustCompile(`Николай II (.+?) (\d{4})(.*)`); 
            m   := re.FindStringSubmatch(cr.coinpage[i].name);
            if (m!=nil) {  
               //fmt.Println(m[1],"----",m[2],"----",m[3]);
               cd[m[1]]=append(cd[m[1]],m[2]); 

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
    fmt.Printf("took %v\n",  time.Since(start))
  
}
