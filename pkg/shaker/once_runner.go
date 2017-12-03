package shaker

import (
	log "github.com/sirupsen/logrus"
	zmq "github.com/pebbe/zmq4"
	"time"
)

type Once struct {
	pullSocket *zmq.Socket
	pushSocket *zmq.Socket
	log        *log.Entry
}

func (runner *Once) Init(config Config, log *log.Entry) error {
	endpoint := config.Master.Socket
	runner.log = log
	pullSocket, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		return err
	}

	pullSocket.Connect(endpoint)
	runner.pullSocket = pullSocket

	go runner.processTasks()

	if !config.IsMaster {
		return nil
	}

	pushSocket, err := zmq.NewSocket(zmq.PUSH)
	pushSocket.Bind(endpoint)
	if err != nil {
		return err
	}

	runner.pushSocket = pushSocket
	return nil
}

func (runner Once) Run(job Job) error {
	_, err := runner.pushSocket.SendMessage(job.Url)
	return err
}

func (runner *Once) Schedule(job Job) {
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

func (runner *Once) processTasks() {
	defer runner.close()

	for {
		msg, err := runner.pullSocket.RecvMessage(0)
		if err != nil {
			runner.log.Errorf("Error: %s", err)
		}

		Fetch(msg[0])
		time.Sleep(time.Second)
	}
}

func (runner *Once) close() (error, error) {
	return runner.pullSocket.Close(), runner.pushSocket.Close()
}
