package instrumentator

import (
	"github.com/sirupsen/logrus"
)

// Instr defines a shareable pre-configured structure containing fields related
// to instrumentator such as Logger and Metric.
type Instr struct {
	Logger *logrus.Entry
}
