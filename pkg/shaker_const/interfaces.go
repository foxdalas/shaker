package shaker

import (
	"github.com/sirupsen/logrus"
)

type Shaker interface {
	Version() string
	Log() *logrus.Entry
}

type Bitbucket interface {
}
