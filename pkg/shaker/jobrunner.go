package shaker

import "github.com/bamzi/jobrunner"

//Run jobs via Jobrunner
func (e RunJob) Run() {
	switch e.request.requestType {
	case "http":
		makeHTTP(e)
	case "redis":
		makeRedis(e)
	}
}

func (s *Shaker) cleanupJobs() {
	for _, job := range jobrunner.Entries() {
		s.Log().Infof("Cleanup job with id %d", job.ID)
		jobrunner.Remove(job.ID)
	}
}
