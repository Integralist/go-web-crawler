package instrumentation

import (
	"github.com/sirupsen/logrus"
)

// Instr defines a shareable pre-configured structure containing fields related
// to instrumentation such as Logger and Metric.
type Instr struct {
	Logger *logrus.Entry
}
