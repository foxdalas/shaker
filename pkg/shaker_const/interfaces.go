package shaker

import (
	"github.com/sirupsen/logrus"
)

//Shaker interface
type Shaker interface {
	Version() string
	Log() *logrus.Entry
}
