package job

import (
	"fmt"
	"net/url"
	"os"
	"time"

	config "github.com/bench-routes/bench-routes/src/lib/config_v2"
	"github.com/bench-routes/bench-routes/src/lib/modules/evaluate"
	"github.com/bench-routes/bench-routes/tsdb/file"
)

type machineJob struct {
	JobInfo
	app   file.Appendable
	sigCh chan struct{}
	host  string
}

func newMachineJob(app file.Appendable, c chan struct{}, api *config.API) (*machineJob, error) {
	newjob := &machineJob{
		app:   app,
		sigCh: c,
		JobInfo: JobInfo{
			Name:        api.Name,
			Every:       api.Every,
			lastExecute: time.Now().Add(api.Every * -1),
		},
	}
	url, err := url.Parse(api.Domain)
	if err != nil {
		return nil, err
	}
	newjob.host = url.Host
	return newjob, nil
}

// Execute execute the machineJob.
func (mn *machineJob) Execute() {
	for range mn.sigCh {
		mn.JobInfo.writeTime()
		ping, jitter, err := evaluate.Machine(mn.host)
		if err != nil {
			fmt.Fprintf(os.Stderr, "job: %s: error: %s", mn.JobInfo.Name, err.Error())
		}
		pingval := fmt.Sprintf("%v|%v|%v", ping.Max.Microseconds(), ping.Mean.Microseconds(), ping.Mean.Microseconds())
		jitterval := fmt.Sprintf("%v", jitter.Value.Microseconds())
		mn.app.Append(file.NewBlock("ping", pingval))
		mn.app.Append(file.NewBlock("jitter", jitterval))
	}
}

// Abort aborts the running job.
func (mn *machineJob) Abort() {
	close(mn.sigCh)
}

// Info returns the jobInfo of the job.
func (mn *machineJob) Info() *JobInfo {
	return &mn.JobInfo
}
