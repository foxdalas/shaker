package shaker

func main() {
	container := CreateRunnerContainer(config, log)
	for job := range (jobs) {
		runner := container[job.Runner]
		go runner.Schedule(job)
	}
}
