package shaker

import (
	log "github.com/sirupsen/logrus"
	zmq "github.com/pebbe/zmq4"
	"time"
)

type SingleRunner struct {
	endpoint   string
	pullSocket *zmq.Socket
	pushSocket *zmq.Socket
	log        *log.Entry
}

func (runner *SingleRunner) Init(config Config, log *log.Entry) error {
	runner.endpoint = config.Master.Socket
	runner.log = log
	context, err := zmq.NewContext()
	if err != nil {
		return err
	}

	pullSocket, err := context.NewSocket(zmq.PULL)
	if err != nil {
		return err
	}

	pullSocket.Connect(runner.endpoint)
	runner.pullSocket = pullSocket

	go runner.processTasks()

	if !config.IsMaster {
		//runner.startHeartbeat()
		return nil
	}

	pushSocket, err := context.NewSocket(zmq.PUSH)
	pushSocket.Bind(runner.endpoint)
	if err != nil {
		return err
	}

	runner.pushSocket = pushSocket
	return nil
}

func (runner SingleRunner) startHeartbeat() {
	//runner.endpoint
}

func (runner SingleRunner) Run(job Job) error {
	_, err := runner.pushSocket.SendMessage(job.Url)
	return err
}

func (runner *SingleRunner) Schedule(job Job) {
	ticker := time.NewTicker(job.Duration)
	go func() {
		for range ticker.C {
			err := runner.Run(job)
			if err != nil {
				runner.log.Errorf("Error: %s", err)
			}
		}
	}()
}

func (runner *SingleRunner) processTasks() {
	defer runner.close()

	for {
		msg, zmqErr := runner.pullSocket.RecvMessage(0)
		if zmqErr != nil {
			runner.log.Errorf("Error: %s", zmqErr)
		}

		fetchErr := Fetch(msg[0])
		if fetchErr != nil {
			runner.log.Errorf("Error: %s", fetchErr)
		}

		time.Sleep(time.Second)
	}
}

func (runner *SingleRunner) close() (error, error) {
	return runner.pullSocket.Close(), runner.pushSocket.Close()
}
