## OpenAir XML API Generator

Use `go generate` to generate an API for the [OpenAir XML API](https://www.openair.com/download/NetSuiteOpenAirXMLAPIGuide.pdf).

### Usage

* `go get -u github.com/joefitzgerald/openair`
* Create `definition.go` in a package that you wish to have generated files in, with the following content:
```
package openair

//go:generate openair -prefix=openair_ -suffix= -object=Customer,Project,User,Timetype,Timesheet,TaskTimecard,Task
```
* Run `go generate .` in the package that contains `definition.go`
* Observe new files generated:
  * `openair_common.go`
  * `openair_customer.go`
  * `openair_project.go`
  * `openair_task.go`
  * `openair_tasktimecard.go`
  * `openair_timesheet.go`
  * `openair_timetype.go`
  * `openair_user.go`

### License

Apache 2.0
