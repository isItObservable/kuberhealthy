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
   "fmt"
  "log"
 "strings"
  "os"
  "os/exec"
  "github.com/kuberhealthy/kuberhealthy/v2/pkg/checks/external/checkclient"
)

func createDatastore() bool
{
    tracetestserver:=os.Getenv("TRACETEST_URL")
    cmd := exec.Command("tracetest", "datastore","apply","-f","/test/datastore.yaml")
    stdout, err := cmd.StdoutPipe()

    if err != nil {
        log.Fatal(err)
        return err
    }

    if err := cmd.Start(); err != nil {
        log.Fatal(err)
        return err
    }

    data, err := ioutil.ReadAll(stdout)

    if err != nil {
        log.Fatal(err)
        return err
    }

    if err := cmd.Wait(); err != nil {
        log.Fatal(err)
        return err
    }

    return true
}
func configureTraceTest() bool
{
    tracetestserver:=os.Getenv("TRACETEST_URL")
    cmd := exec.Command("tracetest", "configure","--endpoint",tracetestserver,"--analytics=false")
    stdout, err := cmd.StdoutPipe()

    if err != nil {
        log.Fatal(err)
        return false
    }

    if err := cmd.Start(); err != nil {
        log.Fatal(err)
        return false
    }

    data, err := ioutil.ReadAll(stdout)

    if err != nil {
        log.Fatal(err)
        return false
    }

    if err := cmd.Wait(); err != nil {
        log.Fatal(err)
        return false
    }



   return true
}

func runTraceTest() bool
{

        cmd := exec.Command("tracetest", "test","run","--definition","/test/test.yaml","--wait-for-result")
        stdout, err := cmd.StdoutPipe()

        if err != nil {
            log.Fatal(err)
            checkclient.ReportFailure([]string{err})
            return false
        }

        if err := cmd.Start(); err != nil {
           log.Fatal(err)
           checkclient.ReportFailure([]string{err})
           return false

        }

        data, err := ioutil.ReadAll(stdout)

        if err != nil {
           log.Fatal(err)
           checkclient.ReportFailure([]string{err})
           return false
        }

        if err := cmd.Wait(); err != nil {
            log.Fatal(err)
            checkclient.ReportFailure([]string{err})
            return false
        }

        check:= strings.Contains(data,"ERROR")
        if check
        {
            log.Fatal(data)
            checkclient.ReportFailure([]string{data})
            return false
        }
        else
            return true
}

func main() {
  if(configureTraceTest())
  {
    if(createDatastore())
    {
        ok := runTraceTest()
        if ok {
            checkclient.ReportSuccess()
        }

    }
        else
            checkclient.ReportFailure([]string{"Unable to Create Datastore"})
  }
  else
    checkclient.ReportFailure([]string{"Unable to configure Tracetest"})

}