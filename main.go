package main

import (
    "bytes"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    //"os"
    "strconv"
    "strings"

    "github.com/alexflint/go-arg"
    "github.com/kr/pretty"
    "github.com/phin1x/go-ipp"
)

func Main() error {
    var args struct {
        Operation      string `arg:"positional,required"`
        URI            string `arg:"positional"`
        PostscriptFile string `arg:"positional"`
        Firststr string `arg:"positional"`
        Secondstr string `arg:"positional"`
    }

    arg.MustParse(&args)

    // define a ipp request
    var req *ipp.Request = nil

    if (args.Operation == "print") {
      req = ipp.NewRequest(ipp.OperationPrintJob, 1)
    } else if (args.Operation == "query") {
      req = ipp.NewRequest(ipp.OperationGetPrinterAttributes, 1)
    } else {
        return fmt.Errorf("Invalid Operation")
    }
    req.OperationAttributes[ipp.AttributeCharset] = "utf-8"
    req.OperationAttributes[ipp.AttributeNaturalLanguage] = "en"
    req.OperationAttributes[ipp.AttributePrinterURI] = args.URI
    req.OperationAttributes[ipp.AttributeRequestingUserName] = "some-user"
    req.OperationAttributes[ipp.AttributeDocumentFormat] = "application/octet-stream"

    // encode request to bytes
    payload, err := req.Encode()
    if err != nil {
        return fmt.Errorf("error encoding ipp request: %w", err)
    }

    // read the test page
    postscript, err := ioutil.ReadFile(args.PostscriptFile)
    if err != nil {
        return fmt.Errorf("error reading postscript file: %w", err)
    }

    // Apply Substitutions
    str1 := strings.Replace(args.Firststr,"(","\\(",-1)
    str1 = strings.Replace(str1,")","\\)",-1)
    str2 := strings.Replace(args.Secondstr,"(","\\(",-1)
    str2 = strings.Replace(str2,")","\\)",-1)

    ps := strings.Replace(string(postscript),"{FIRSTSTRING}",str1,-1)
    ps = strings.Replace(ps,"{SECONDSTRING}",str2,-1)
    payload = append(payload, postscript...)

    // send ipp request to remote server via http
    httpReq, err := http.NewRequest("POST", args.URI, bytes.NewReader(payload))
    if err != nil {
        return fmt.Errorf("error creating http request: %w", err)
    }

    // set ipp headers
    httpReq.Header.Set("Content-Length", strconv.Itoa(len(payload)))
    httpReq.Header.Set("Content-Type", ipp.ContentTypeIPP)

    // perform the request
    var httpClient http.Client
    httpResp, err := httpClient.Do(httpReq)
    if err != nil {
        return fmt.Errorf("error executing http request: %w", err)
    }
    defer httpResp.Body.Close()

    // read the response
    buf, err := io.ReadAll(httpResp.Body)
    if err != nil {
        return fmt.Errorf("error reading response body: %w", err)
    }

    //fmt.Println (buf)

    // response must be 200 for a successful operation
    // other possible http codes are:
    // - 500 -> server error
    // - 426 -> sever requests a encrypted connection
    // - 401 -> forbidden -> need authorization header or user is not permitted
    if httpResp.StatusCode != 200 {
        return fmt.Errorf("printer said %d: %s", httpResp.StatusCode, buf)
    }

    // decode ipp response
    resp, err := ipp.NewResponseDecoder(bytes.NewReader(buf)).Decode(nil)
    if err != nil {
        return fmt.Errorf("error decoding ipp response: %w", err)
    }

    // print the response
    fmt.Println("Submitted print job. Response was:")
    pretty.Println(resp)
    return nil
}

func main () {
  fmt.Println("Main")
  err := Main() 
  fmt.Println("Returned",err)
}
