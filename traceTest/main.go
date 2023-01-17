// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
  "strings"
  "os"
  "bytes"
  "fmt"
  "os/exec"
  "io/ioutil"
  "github.com/kuberhealthy/kuberhealthy/v2/pkg/checks/external/checkclient"
)

func createDatastore() bool {
     cmd := exec.Command("tracetest", "datastore","apply","-f","/test/datastore.yaml")
     var out bytes.Buffer
     var stderr bytes.Buffer
     cmd.Stdout = &out
     cmd.Stderr = &stderr

     err := cmd.Run()
     if err != nil {
         fmt.Println("Error "+ fmt.Sprint(err) + ": " + stderr.String())
         return false
     }

    fmt.Println("output datastore :" +out.String())
    return true
}


func checkTraceTest() bool {

    cmd := exec.Command("cat", "/test/test.yaml")
    stdout, err := cmd.StdoutPipe()

    if err != nil {
        fmt.Println(err.Error())
        return false
    }

    if err := cmd.Start(); err != nil {
        fmt.Println("Error:"+ err.Error())
        return false
    }

    data, err := ioutil.ReadAll(stdout)
    datastring := string(data)
    if err != nil {
        fmt.Println("Error:"+err.Error())
        return false
    }

    if err := cmd.Wait(); err != nil {
        fmt.Println("Error:"+err.Error())
        return false
    }
    fmt.Println("definition file to run :" +datastring)

   return true
}

func configureTraceTest() bool {
    tracetestserver:=os.Getenv("TRACETEST_URL")
    fmt.Println("Tracetest url :"+tracetestserver)
    cmd := exec.Command("tracetest", "configure","--endpoint",tracetestserver,"--analytics=false")
    stdout, err := cmd.Output()


    if err != nil {
        fmt.Println("Error:"+err.Error())
        return false
    }
    datastring := string(stdout)


    fmt.Println("output configure:" +datastring)

   return true
}

func runTraceTest() bool {

        cmd := exec.Command("tracetest", "test","run","-d","/test/test.yaml","--wait-for-result")

        stdout, err := cmd.StdoutPipe()


        if err != nil {
             fmt.Println("Error: "+ fmt.Sprint(err)  )
        }

        if err := cmd.Start(); err != nil {
            fmt.Println("Error start: "+ fmt.Sprint(err) )
        }

        data, err := ioutil.ReadAll(stdout)

        if err != nil {
            fmt.Println("Error reading : "+ fmt.Sprint(err) )
        }

        if err := cmd.Wait(); err != nil {
            fmt.Println("Error wainting : "+ fmt.Sprint(err) )
        }



        datastring := string(data)
        fmt.Println("output run:" +datastring)
        check:= strings.Contains(datastring,"ERROR")
        if check {
            fmt.Println("outptut contains error: "+datastring)
            checkclient.ReportFailure([]string{datastring})
            return false
        } else {
            return true
        }
}

func main() {
  if checkTraceTest() {
      fmt.Println("Check with success")
      if configureTraceTest() {
        fmt.Println("Configure with success")
      //  if createDatastore()  {
        //    fmt.Println("Create datastore with success")
            ok := runTraceTest()
            if ok {
                fmt.Println("run trace test with success")
                checkclient.ReportSuccess()
            } else
            {
               fmt.Println("run trace test with errors")
            }

        //} else {
          //      fmt.Println("Unable to create Datastore")
           //     checkclient.ReportFailure([]string{"Unable to Create Datastore"})
        //}
      } else {
         fmt.Println("Unable to configure tracetest" )
        checkclient.ReportFailure([]string{"Unable to configure Tracetest"})
      }
  }  else {
        fmt.Println("Issue calling tracetest")
  }

}